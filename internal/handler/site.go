package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/db"
	"waxp/echo/internal/filter"
)

type SiteHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewSiteHandler(queries *db.Queries, pool *pgxpool.Pool) *SiteHandler {
	return &SiteHandler{queries: queries, pool: pool}
}

type CreateSiteRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=255"`
	Domain string `json:"domain" validate:"required,min=1,max=255"`
}

type UpdateSiteRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=255"`
	Domain string `json:"domain" validate:"required,min=1,max=255"`
}

type SiteResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor *int64 `json:"next_cursor"`
	Total      int64  `json:"total"`
	HasMore    bool   `json:"has_more"`
}

func (h *SiteHandler) Create(c *echo.Context) error {
	var req CreateSiteRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Domain == "" {
		return ErrorJSON(c, http.StatusBadRequest, "name and domain are required")
	}

	site, err := h.queries.CreateSite(c.Request().Context(), db.CreateSiteParams{
		Name:   req.Name,
		Domain: req.Domain,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "domain already exists")
		}
		return ErrorJSON(c, http.StatusInternalServerError, "failed to create site")
	}

	return c.JSON(http.StatusCreated, SiteResponse{
		ID:     site.ID,
		Name:   site.Name,
		Domain: site.Domain,
	})
}

func (h *SiteHandler) GetByID(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid id")
	}

	site, err := h.queries.GetSiteByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return ErrorJSON(c, http.StatusInternalServerError, "failed to get site")
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:     site.ID,
		Name:   site.Name,
		Domain: site.Domain,
	})
}

func (h *SiteHandler) List(c *echo.Context) error {
	const maxLimit int32 = 100

	limitStr := c.QueryParam("limit")
	cursorStr := c.QueryParam("cursor")

	var limit *int32
	if limitStr != "" {
		parsed, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil || parsed <= 0 {
			return ErrorJSON(c, http.StatusBadRequest, "invalid limit")
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
			return ErrorJSON(c, http.StatusBadRequest, "invalid cursor")
		}
		cursor = &parsed
	}

	builder := filter.NewBuilder(map[string]string{
		"name":   "name",
		"domain": "domain",
		"id":     "id",
	})
	if err := builder.Parse(c.Request().URL.Query()); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, err.Error())
	}

	result := builder.Build(cursor)
	ctx := c.Request().Context()

	countSQL := "SELECT COUNT(*) FROM sites" + result.WhereClause
	var total int64
	if err := h.pool.QueryRow(ctx, countSQL, result.Args...).Scan(&total); err != nil {
		return ErrorJSON(c, http.StatusInternalServerError, "failed to count sites")
	}

	listSQL := "SELECT id, name, domain, created_at, updated_at FROM sites" + result.WhereClause + " ORDER BY id ASC"

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
		return ErrorJSON(c, http.StatusInternalServerError, "failed to list sites")
	}
	defer rows.Close()

	var sites []db.Site
	for rows.Next() {
		var s db.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return ErrorJSON(c, http.StatusInternalServerError, "failed to scan site")
		}
		sites = append(sites, s)
	}
	if err := rows.Err(); err != nil {
		return ErrorJSON(c, http.StatusInternalServerError, "failed to list sites")
	}

	data := make([]SiteResponse, 0, len(sites))
	for _, s := range sites {
		data = append(data, SiteResponse{
			ID:     s.ID,
			Name:   s.Name,
			Domain: s.Domain,
		})
	}

	var nextCursor *int64
	var hasMore bool
	if paginated {
		hasMore = len(sites) == int(*limit)
		if hasMore && len(sites) > 0 {
			lastID := sites[len(sites)-1].ID
			nextCursor = &lastID
		}
	}

	return c.JSON(http.StatusOK, PaginatedResponse[SiteResponse]{
		Data:       data,
		NextCursor: nextCursor,
		Total:      total,
		HasMore:    hasMore,
	})
}

func (h *SiteHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateSiteRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Domain == "" {
		return ErrorJSON(c, http.StatusBadRequest, "name and domain are required")
	}

	site, err := h.queries.UpdateSite(c.Request().Context(), db.UpdateSiteParams{
		Name:   req.Name,
		Domain: req.Domain,
		ID:     id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "domain already exists")
		}
		return ErrorJSON(c, http.StatusInternalServerError, "failed to update site")
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:     site.ID,
		Name:   site.Name,
		Domain: site.Domain,
	})
}

func (h *SiteHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid id")
	}

	_, err = h.queries.GetSiteByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return ErrorJSON(c, http.StatusInternalServerError, "failed to get site")
	}

	if err := h.queries.DeleteSite(c.Request().Context(), id); err != nil {
		return ErrorJSON(c, http.StatusInternalServerError, "failed to delete site")
	}

	return c.JSON(http.StatusNoContent, nil)
}
