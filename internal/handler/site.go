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

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
	"waxp/echo/internal/filter"
	"waxp/echo/internal/i18n"
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
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Domain == "" {
		return apierror.JSON(c, http.StatusBadRequest, "name and domain are required")
	}

	if len(req.Name) > 255 {
		return apierror.JSON(c, http.StatusBadRequest, "name must be at most 255 characters")
	}

	if err := validateDomain(req.Domain); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	if len(req.Locales) == 0 {
		return apierror.JSON(c, http.StatusBadRequest, "at least one locale is required")
	}

	defaultCount := 0
	seenLocales := make(map[string]bool, len(req.Locales))
	for _, l := range req.Locales {
		if l.Code == "" {
			return apierror.JSON(c, http.StatusBadRequest, "locale code is required")
		}
		if err := validateLocaleCode(l.Code); err != nil {
			return apierror.JSON(c, http.StatusBadRequest, err.Error())
		}
		if seenLocales[l.Code] {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("duplicate locale code '%s'", l.Code))
		}
		seenLocales[l.Code] = true
		if l.IsDefault {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return apierror.JSON(c, http.StatusBadRequest, "only one locale can be the default")
	}

	options := req.Options
	if options == nil {
		options = json.RawMessage(`{}`)
	}

	if err := validateJSON(options); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "options: "+err.Error())
	}

	ctx := c.Request().Context()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return apierror.Internal(c, "failed to begin transaction", err)
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
			return apierror.JSON(c, http.StatusConflict, "domain already exists")
		}
		return apierror.Internal(c, "failed to create site", err)
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
				return apierror.JSON(c, http.StatusConflict, fmt.Sprintf("locale code '%s' already exists", l.Code))
			}
			return apierror.Internal(c, "failed to create locale", err)
		}
		dbLocales = append(dbLocales, locale)
	}

	now := time.Now().UTC()
	publishedAt := pgtype.Timestamptz{Time: now, Valid: true}

	_, err = qtx.CreateSectionCounter(ctx, site.ID)
	if err != nil {
		return apierror.Internal(c, "failed to create section counter", err)
	}

	_, err = qtx.CreateBlockCounter(ctx, site.ID)
	if err != nil {
		return apierror.Internal(c, "failed to create block counter", err)
	}

	makeLayout := func() ([]byte, error) {
		return makeDefaultLayout(func() (int64, error) {
			return qtx.GetNextSectionID(ctx, site.ID)
		}, 4)
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
			return apierror.Internal(c, "failed to generate section id", err)
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
			return apierror.Internal(c, "failed to create default page", err)
		}

		for _, loc := range dbLocales {
			_, err := qtx.CreatePageSlug(ctx, db.CreatePageSlugParams{
				PageID:   page.ID,
				LocaleID: loc.ID,
				Slug:     dp.slug,
			})
			if err != nil {
				return apierror.Internal(c, "failed to create default page slug", err)
			}

			_, err = qtx.CreatePageSeo(ctx, db.CreatePageSeoParams{
				PageID:      page.ID,
				LocaleID:    loc.ID,
				Title:       dp.title,
				Description: pgtype.Text{},
			})
			if err != nil {
				return apierror.Internal(c, "failed to create default page seo", err)
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
		return apierror.Internal(c, "failed to commit transaction", err)
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
	id, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	locale := c.QueryParam("locale")

	ctx := c.Request().Context()

	site, err := h.queries.GetSiteByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	locales, err := h.queries.ListSiteLocales(ctx, id)
	if err != nil {
		return apierror.Internal(c, "failed to get locales", err)
	}

	if locale != "" {
		validLocale := false
		for _, l := range locales {
			if l.Code == locale {
				validLocale = true
			}
		}
		if !validLocale {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale '%s' does not belong to this site", locale))
		}
	} else {
		for _, l := range locales {
			if l.IsDefault {
				locale = l.Code
				break
			}
		}
	}

	routes, err := buildRoutesMap(ctx, h.queries, id)
	if err != nil {
		return err
	}

	resolvedOptions, err := i18n.Resolve(site.Options, locale)
	if err != nil {
		return apierror.Internal(c, "failed to resolve options locales", err)
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: resolvedOptions,
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
		"name":   "name",
		"domain": "domain",
		"id":     "id",
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

	countSQL := "SELECT COUNT(*) FROM sites" + whereClause
	var total int64
	if err := h.pool.QueryRow(ctx, countSQL, result.Args...).Scan(&total); err != nil {
		return apierror.Internal(c, "failed to count sites", err)
	}

	listSQL := "SELECT id, name, domain, options, created_at, updated_at, is_live FROM sites" + whereClause + " ORDER BY id ASC"

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
		return apierror.Internal(c, "failed to list sites", err)
	}
	defer rows.Close()

	var sites []db.Site
	for rows.Next() {
		var s db.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.Options, &s.CreatedAt, &s.UpdatedAt, &s.IsLive); err != nil {
			return apierror.Internal(c, "failed to scan site", err)
		}
		sites = append(sites, s)
	}
	if err := rows.Err(); err != nil {
		return apierror.Internal(c, "failed to list sites", err)
	}

	siteIDs := make([]int64, len(sites))
	for i, s := range sites {
		siteIDs[i] = s.ID
	}

	var localesBySite map[int64][]db.SiteLocale
	if len(siteIDs) > 0 {
		allLocales, err := h.queries.ListSiteLocalesBySiteIDs(ctx, siteIDs)
		if err != nil {
			return apierror.Internal(c, "failed to get locales", err)
		}
		localesBySite = make(map[int64][]db.SiteLocale, len(siteIDs))
		for _, l := range allLocales {
			localesBySite[l.SiteID] = append(localesBySite[l.SiteID], l)
		}
	}

	data := make([]SiteResponse, 0, len(sites))
	for _, s := range sites {
		locales := localesBySite[s.ID]
		if locales == nil {
			locales = []db.SiteLocale{}
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
	id, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	locale := c.QueryParam("locale")
	if locale == "" {
		return apierror.JSON(c, http.StatusBadRequest, "locale is required")
	}

	var req UpdateSiteRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.Domain == "" {
		return apierror.JSON(c, http.StatusBadRequest, "name and domain are required")
	}

	if len(req.Name) > 255 {
		return apierror.JSON(c, http.StatusBadRequest, "name must be at most 255 characters")
	}

	if err := validateDomain(req.Domain); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	existing, err := h.queries.GetSiteByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	siteLocales, err := h.queries.ListSiteLocales(ctx, id)
	if err != nil {
		return apierror.Internal(c, "failed to get locales", err)
	}

	validLocale := false
	for _, l := range siteLocales {
		if l.Code == locale {
			validLocale = true
		}
	}
	if !validLocale {
		return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale '%s' does not belong to this site", locale))
	}

	options := req.Options
	if options == nil {
		options = existing.Options
	} else {
		if err := validateJSON(options); err != nil {
			return apierror.JSON(c, http.StatusBadRequest, "options: "+err.Error())
		}
		options, err = i18n.Merge(options, existing.Options, locale)
		if err != nil {
			return apierror.Internal(c, "failed to merge options locales", err)
		}
	}

	site, err := h.queries.UpdateSite(ctx, db.UpdateSiteParams{
		Name:    req.Name,
		Domain:  req.Domain,
		Options: options,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierror.JSON(c, http.StatusConflict, "domain already exists")
		}
		return apierror.Internal(c, "failed to update site", err)
	}

	resolvedOptions, err := i18n.Resolve(site.Options, locale)
	if err != nil {
		return apierror.Internal(c, "failed to resolve options locales", err)
	}

	return c.JSON(http.StatusOK, SiteResponse{
		ID:      site.ID,
		Name:    site.Name,
		Domain:  site.Domain,
		Options: resolvedOptions,
		Locales: toLocaleResponses(siteLocales),
	})
}

func (h *SiteHandler) Delete(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	_, err = h.queries.DeleteSite(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to delete site", err)
	}

	return c.NoContent(http.StatusNoContent)
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
