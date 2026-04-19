package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
	"waxp/echo/internal/filter"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

type MediaHandler struct {
	queries  *db.Queries
	pool     *pgxpool.Pool
	mediaDir string
}

func NewMediaHandler(queries *db.Queries, pool *pgxpool.Pool, mediaDir string) *MediaHandler {
	return &MediaHandler{queries: queries, pool: pool, mediaDir: mediaDir}
}

type MediaResponse struct {
	ID        int64  `json:"id"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

func (h *MediaHandler) Upload(c *echo.Context) error {
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "file is required")
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("mime type '%s' not allowed", mimeType))
	}

	if err := os.MkdirAll(h.mediaDir, 0755); err != nil {
		return apierror.Internal(c, "failed to create media directory", err)
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch mimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "image/svg+xml":
			ext = ".svg"
		}
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dstPath := filepath.Join(h.mediaDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return apierror.Internal(c, "failed to create file", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(dstPath)
		return apierror.Internal(c, "failed to save file", err)
	}

	url := "/media/" + filename

	media, err := h.queries.CreateMedia(c.Request().Context(), db.CreateMediaParams{
		Filename: header.Filename,
		MimeType: mimeType,
		Size:     size,
		Url:      url,
	})
	if err != nil {
		os.Remove(dstPath)
		return apierror.Internal(c, "failed to save media record", err)
	}

	return c.JSON(http.StatusCreated, toMediaResponse(media))
}

func (h *MediaHandler) List(c *echo.Context) error {
	const maxLimit int32 = 100

	limitStr := c.QueryParam("limit")
	cursorStr := c.QueryParam("cursor")

	var limit *int32
	if limitStr != "" {
		parsed, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || parsed <= 0 {
			return apierror.JSON(c, http.StatusBadRequest, "invalid limit")
		}
		v := int32(parsed)
		if v > maxLimit {
			v = maxLimit
		}
		limit = &v
	}

	var cursor *int64
	if cursorStr != "" {
		parsed, err := strconv.ParseInt(cursorStr, 10, 64)
		if err != nil || parsed < 0 {
			return apierror.JSON(c, http.StatusBadRequest, "invalid cursor")
		}
		cursor = &parsed
	}

	builder := filter.NewBuilder(map[string]string{
		"id":        "id",
		"filename":  "filename",
		"mime_type": "mime_type",
	})
	if err := builder.Parse(c.Request().URL.Query()); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	result := builder.Build(cursor, 1)
	ctx := c.Request().Context()

	whereClause := ""
	if result.WhereClause != "" {
		whereClause = " WHERE " + result.WhereClause
	}

	countSQL := "SELECT COUNT(*) FROM media" + whereClause
	var total int64
	if err := h.pool.QueryRow(ctx, countSQL, result.Args...).Scan(&total); err != nil {
		return apierror.Internal(c, "failed to count media", err)
	}

	listSQL := "SELECT id, filename, mime_type, size, url, created_at FROM media" + whereClause + " ORDER BY id ASC"

	paginated := limit != nil
	var listArgs []any
	if paginated {
		nextParam := len(result.Args) + 1
		listSQL += fmt.Sprintf(" LIMIT $%d", nextParam)
		listArgs = append(append([]any{}, result.Args...), int64(*limit))
	} else {
		listArgs = result.Args
	}

	rows, err := h.pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return apierror.Internal(c, "failed to list media", err)
	}
	defer rows.Close()

	var media []db.Medium
	for rows.Next() {
		var m db.Medium
		if err := rows.Scan(&m.ID, &m.Filename, &m.MimeType, &m.Size, &m.Url, &m.CreatedAt); err != nil {
			return apierror.Internal(c, "failed to scan media", err)
		}
		media = append(media, m)
	}
	if err := rows.Err(); err != nil {
		return apierror.Internal(c, "failed to list media", err)
	}

	data := make([]MediaResponse, 0, len(media))
	for _, m := range media {
		data = append(data, toMediaResponse(m))
	}

	var nextCursor *int64
	var hasMore bool
	if paginated {
		hasMore = len(media) == int(*limit)
		if hasMore && len(media) > 0 {
			lastID := media[len(media)-1].ID
			nextCursor = &lastID
		}
	}

	return c.JSON(http.StatusOK, PaginatedResponse[MediaResponse]{
		Data:       data,
		NextCursor: nextCursor,
		Total:      total,
		HasMore:    hasMore,
	})
}

func (h *MediaHandler) GetByID(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	media, err := h.queries.GetMediaByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "media not found")
		}
		return apierror.Internal(c, "failed to get media", err)
	}

	return c.JSON(http.StatusOK, toMediaResponse(media))
}

func (h *MediaHandler) Delete(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	media, err := h.queries.DeleteMedia(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "media not found")
		}
		return apierror.Internal(c, "failed to delete media", err)
	}

	filename := strings.TrimPrefix(media.Url, "/media/")
	os.Remove(filepath.Join(h.mediaDir, filename))

	return c.NoContent(http.StatusNoContent)
}

func ServeMedia(mediaDir string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return apierror.JSON(c, http.StatusBadRequest, "filename is required")
		}

		filePath := filepath.Join(mediaDir, filepath.Base(name))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return apierror.JSON(c, http.StatusNotFound, "file not found")
		}

		return c.File(filePath)
	}
}

func toMediaResponse(m db.Medium) MediaResponse {
	return MediaResponse{
		ID:        m.ID,
		Filename:  m.Filename,
		MimeType:  m.MimeType,
		Size:      m.Size,
		URL:       m.Url,
		CreatedAt: m.CreatedAt.Time.Format(time.RFC3339),
	}
}
