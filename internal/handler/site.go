package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

type CreateSiteLocaleInput struct {
	Code      string `json:"code"`
	IsDefault bool   `json:"is_default"`
}

type CreateSiteRequest struct {
	Name    string                  `json:"name" validate:"required,min=1,max=255"`
	Domain  string                  `json:"domain" validate:"required,min=1,max=255"`
	Options json.RawMessage         `json:"options,omitempty"`
	Locales []CreateSiteLocaleInput `json:"locales,omitempty"`
}

type UpdateSiteRequest struct {
	Name    string          `json:"name" validate:"required,min=1,max=255"`
	Domain  string          `json:"domain" validate:"required,min=1,max=255"`
	Options json.RawMessage `json:"options,omitempty"`
}

type SiteResponse struct {
	ID      int64                   `json:"id"`
	Name    string                  `json:"name"`
	Domain  string                  `json:"domain"`
	Options json.RawMessage         `json:"options"`
	Locales []LocaleResponse        `json:"locales"`
	Routes  map[string][]RouteEntry `json:"routes"`
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

	defaultCount := 0
	for _, l := range req.Locales {
		if l.Code == "" {
			return ErrorJSON(c, http.StatusBadRequest, "locale code is required")
		}
		if l.IsDefault {
			defaultCount++
		}
	}
	if defaultCount > 1 {
		return ErrorJSON(c, http.StatusBadRequest, "only one locale can be the default")
	}

	options := req.Options
	if options == nil {
		options = json.RawMessage(`{}`)
	}

	ctx := c.Request().Context()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return InternalError(c, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	site, err := qtx.CreateSite(ctx, db.CreateSiteParams{
		Name:    req.Name,
		Domain:  req.Domain,
		Options: options,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "domain already exists")
		}
		return InternalError(c, "failed to create site", err)
	}

	var locales []LocaleResponse
	for _, l := range req.Locales {
		locale, err := qtx.CreateSiteLocale(ctx, db.CreateSiteLocaleParams{
			SiteID:    site.ID,
			Code:      l.Code,
			IsDefault: l.IsDefault,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrorJSON(c, http.StatusConflict, fmt.Sprintf("locale code '%s' already exists", l.Code))
			}
			return InternalError(c, "failed to create locale", err)
		}
		locales = append(locales, LocaleResponse{
			Code:      locale.Code,
			IsDefault: locale.IsDefault,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return InternalError(c, "failed to commit transaction", err)
	}

	if locales == nil {
		locales = []LocaleResponse{}
	}

	return c.JSON(http.StatusCreated, SiteResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: site.Options,
		Locales: locales,
	})
}

type DefaultPageResponse struct {
	ID         int64  `json:"id"`
	SiteID     int64  `json:"site_id"`
	Type       string `json:"type"`
	Slug       string `json:"slug"`
	LocaleCode string `json:"locale_code"`
}

type CreateSiteWithDefaultsResponse struct {
	ID      int64                 `json:"id"`
	Name    string                `json:"name"`
	Domain  string                `json:"domain"`
	Options json.RawMessage       `json:"options"`
	Locales []LocaleResponse      `json:"locales"`
	Pages   []DefaultPageResponse `json:"pages"`
}

func (h *SiteHandler) CreateWithDefaults(c *echo.Context) error {
	var req CreateSiteRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Domain == "" {
		return ErrorJSON(c, http.StatusBadRequest, "name and domain are required")
	}

	if len(req.Locales) == 0 {
		return ErrorJSON(c, http.StatusBadRequest, "at least one locale is required")
	}

	defaultCount := 0
	for _, l := range req.Locales {
		if l.Code == "" {
			return ErrorJSON(c, http.StatusBadRequest, "locale code is required")
		}
		if l.IsDefault {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return ErrorJSON(c, http.StatusBadRequest, "only one locale can be the default")
	}

	options := req.Options
	if options == nil {
		options = json.RawMessage(`{}`)
	}

	ctx := c.Request().Context()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return InternalError(c, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	site, err := qtx.CreateSite(ctx, db.CreateSiteParams{
		Name:    req.Name,
		Domain:  req.Domain,
		Options: options,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "domain already exists")
		}
		return InternalError(c, "failed to create site", err)
	}

	var dbLocales []db.SiteLocale
	for _, l := range req.Locales {
		locale, err := qtx.CreateSiteLocale(ctx, db.CreateSiteLocaleParams{
			SiteID:    site.ID,
			Code:      l.Code,
			IsDefault: l.IsDefault,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrorJSON(c, http.StatusConflict, fmt.Sprintf("locale code '%s' already exists", l.Code))
			}
			return InternalError(c, "failed to create locale", err)
		}
		dbLocales = append(dbLocales, locale)
	}

	now := time.Now().UTC()
	publishedAt := pgtype.Timestamptz{Time: now, Valid: true}

	_, err = qtx.CreateSectionCounter(ctx, site.ID)
	if err != nil {
		return InternalError(c, "failed to create section counter", err)
	}

	_, err = qtx.CreateBlockCounter(ctx, site.ID)
	if err != nil {
		return InternalError(c, "failed to create block counter", err)
	}

	makeLayout := func() ([]byte, error) {
		sectionIDs := make([]int64, 4)
		for i := range sectionIDs {
			id, err := qtx.GetNextSectionID(ctx, site.ID)
			if err != nil {
				return nil, err
			}
			sectionIDs[i] = id
		}

		blockIDs := make([]int64, 4)
		for i := range blockIDs {
			id, err := qtx.GetNextBlockID(ctx, site.ID)
			if err != nil {
				return nil, err
			}
			blockIDs[i] = id
		}

		defaultBlock := func(id int64) map[string]interface{} {
			return map[string]interface{}{
				"id":      id,
				"type":    "Text",
				"content": "La mare que va en calsonsillos",
				"d":       map[string]int{"x": 1, "w": 8, "y": 1, "h": 8},
				"t":       map[string]int{"x": 1, "w": 6, "y": 1, "h": 6},
				"m":       map[string]int{"x": 1, "w": 4, "y": 1, "h": 4},
			}
		}

		makeSection := func(id int64, blockID int64) map[string]interface{} {
			return map[string]interface{}{
				"id":      id,
				"mobile":  map[string]int{"cols": 8, "rows": 12, "gap": 4},
				"tablet":  map[string]int{"cols": 20, "rows": 12, "gap": 6},
				"desktop": map[string]int{"cols": 24, "rows": 12, "gap": 8},
				"blocks":  []interface{}{defaultBlock(blockID)},
			}
		}

		defaultLayout := []map[string]interface{}{
			makeSection(sectionIDs[0], blockIDs[0]),
			makeSection(sectionIDs[1], blockIDs[1]),
			makeSection(sectionIDs[2], blockIDs[2]),
			makeSection(sectionIDs[3], blockIDs[3]),
		}
		return json.Marshal(defaultLayout)
	}

	defaultPages := []struct {
		slug  string
		title string
	}{
		{slug: "", title: "Home"},
		{slug: "404", title: "404 page"},
	}

	var pages []DefaultPageResponse
	for _, dp := range defaultPages {
		layout, err := makeLayout()
		if err != nil {
			return InternalError(c, "failed to generate section id", err)
		}

		page, err := qtx.CreatePage(ctx, db.CreatePageParams{
			SiteID:      site.ID,
			BlogID:      pgtype.Int8{},
			ParentID:    pgtype.Int8{},
			Type:        "page",
			Layout:      layout,
			PublishedAt: publishedAt,
		})
		if err != nil {
			return InternalError(c, "failed to create default page", err)
		}

		for _, loc := range dbLocales {
			_, err := qtx.CreatePageSlug(ctx, db.CreatePageSlugParams{
				PageID:   page.ID,
				LocaleID: loc.ID,
				Slug:     dp.slug,
			})
			if err != nil {
				return InternalError(c, "failed to create default page slug", err)
			}

			_, err = qtx.CreatePageSeo(ctx, db.CreatePageSeoParams{
				PageID:      page.ID,
				LocaleID:    loc.ID,
				Title:       dp.title,
				Description: pgtype.Text{},
			})
			if err != nil {
				return InternalError(c, "failed to create default page seo", err)
			}

			pages = append(pages, DefaultPageResponse{
				ID:         page.ID,
				SiteID:     site.ID,
				Type:       "page",
				Slug:       dp.slug,
				LocaleCode: loc.Code,
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return InternalError(c, "failed to commit transaction", err)
	}

	localeResponses := toLocaleResponses(dbLocales)
	if pages == nil {
		pages = []DefaultPageResponse{}
	}

	return c.JSON(http.StatusCreated, CreateSiteWithDefaultsResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: site.Options,
		Locales: localeResponses,
		Pages:   pages,
	})
}

func (h *SiteHandler) GetByID(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid id")
	}

	ctx := c.Request().Context()

	site, err := h.queries.GetSiteByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

	locales, err := h.queries.ListSiteLocales(ctx, id)
	if err != nil {
		return InternalError(c, "failed to get locales", err)
	}

	routes, err := buildRoutesMap(ctx, h.queries, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: site.Options,
		Locales: toLocaleResponses(locales),
		Routes:  routes,
	})
}

func buildRoutesMap(ctx context.Context, q *db.Queries, siteID int64) (map[string][]RouteEntry, error) {
	routes := make(map[string][]RouteEntry)

	pageRoutes, err := q.GetPageRoutes(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get page routes: %w", err)
	}
	for _, r := range pageRoutes {
		path := buildRoutePath(r.LocaleCode, r.IsDefault, r.Path)
		routes[r.LocaleCode] = append(routes[r.LocaleCode], RouteEntry{
			Path:   path,
			PageID: &r.PageID,
		})
	}

	blogRoutes, err := q.GetBlogRoutes(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blog routes: %w", err)
	}
	for _, r := range blogRoutes {
		path := buildRoutePath(r.LocaleCode, r.IsDefault, r.Path)
		routes[r.LocaleCode] = append(routes[r.LocaleCode], RouteEntry{
			Path:   path,
			BlogID: &r.BlogID,
		})
	}

	postRoutes, err := q.GetPostRoutes(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post routes: %w", err)
	}
	for _, r := range postRoutes {
		path := buildRoutePath(r.LocaleCode, r.IsDefault, r.Path)
		routes[r.LocaleCode] = append(routes[r.LocaleCode], RouteEntry{
			Path:   path,
			PageID: &r.PageID,
		})
	}

	return routes, nil
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
		return InternalError(c, "failed to count sites", err)
	}

	listSQL := "SELECT id, name, domain, options, created_at, updated_at FROM sites" + result.WhereClause + " ORDER BY id ASC"

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
		return InternalError(c, "failed to list sites", err)
	}
	defer rows.Close()

	var sites []db.Site
	for rows.Next() {
		var s db.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.Options, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return InternalError(c, "failed to scan site", err)
		}
		sites = append(sites, s)
	}
	if err := rows.Err(); err != nil {
		return InternalError(c, "failed to list sites", err)
	}

	data := make([]SiteResponse, 0, len(sites))
	for _, s := range sites {
		locales, err := h.queries.ListSiteLocales(ctx, s.ID)
		if err != nil {
			return InternalError(c, "failed to get locales", err)
		}
		data = append(data, SiteResponse{
			ID:      s.ID,
			Name:    s.Name,
			Domain:  s.Domain,
			Options: s.Options,
			Locales: toLocaleResponses(locales),
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

	options := req.Options
	if options == nil {
		options = json.RawMessage(`{}`)
	}

	site, err := h.queries.UpdateSite(c.Request().Context(), db.UpdateSiteParams{
		Name:    req.Name,
		Domain:  req.Domain,
		Options: options,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "domain already exists")
		}
		return InternalError(c, "failed to update site", err)
	}

	locales, err := h.queries.ListSiteLocales(c.Request().Context(), id)
	if err != nil {
		return InternalError(c, "failed to get locales", err)
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: site.Options,
		Locales: toLocaleResponses(locales),
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
		return InternalError(c, "failed to get site", err)
	}

	if err := h.queries.DeleteSite(c.Request().Context(), id); err != nil {
		return InternalError(c, "failed to delete site", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

func toLocaleResponses(locales []db.SiteLocale) []LocaleResponse {
	result := make([]LocaleResponse, 0, len(locales))
	for _, l := range locales {
		result = append(result, LocaleResponse{
			Code:      l.Code,
			IsDefault: l.IsDefault,
		})
	}
	return result
}
