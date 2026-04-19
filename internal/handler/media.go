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
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
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
	mediaDir string
}

func NewMediaHandler(queries *db.Queries, mediaDir string) *MediaHandler {
	return &MediaHandler{queries: queries, mediaDir: mediaDir}
}

type MediaResponse struct {
	ID        int64  `json:"id"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type MediaListResponse struct {
	Data    []MediaResponse `json:"data"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
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
	page := 1
	perPage := 20

	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.QueryParam("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}

	total, err := h.queries.CountMedia(c.Request().Context())
	if err != nil {
		return apierror.Internal(c, "failed to count media", err)
	}

	offset := (page - 1) * perPage
	media, err := h.queries.ListMedia(c.Request().Context(), db.ListMediaParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return apierror.Internal(c, "failed to list media", err)
	}

	data := make([]MediaResponse, 0, len(media))
	for _, m := range media {
		data = append(data, toMediaResponse(m))
	}

	return c.JSON(http.StatusOK, MediaListResponse{
		Data:    data,
		Total:   total,
		Page:    page,
		PerPage: perPage,
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
