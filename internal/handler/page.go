package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
	"waxp/echo/internal/filter"
)

type PageHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewPageHandler(queries *db.Queries, pool *pgxpool.Pool) *PageHandler {
	return &PageHandler{queries: queries, pool: pool}
}

type PageSlugInput struct {
	LocaleCode string `json:"locale_code"`
	Slug       string `json:"slug"`
}

type PageSeoInput struct {
	LocaleCode  string `json:"locale_code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreatePageRequest struct {
	Type        string          `json:"type"`
	BlogID      *int64          `json:"blog_id"`
	ParentID    *int64          `json:"parent_id"`
	Layout      json.RawMessage `json:"layout"`
	PublishedAt *string         `json:"published_at"`
	Seo         []PageSeoInput  `json:"seo"`
	Slugs       []PageSlugInput `json:"slugs"`
}

type UpdatePageRequest struct {
	ParentID    *int64          `json:"parent_id"`
	Layout      json.RawMessage `json:"layout"`
	PublishedAt *string         `json:"published_at"`
	Seo         []PageSeoInput  `json:"seo"`
	Slugs       []PageSlugInput `json:"slugs"`
}

type PageSlugResponse struct {
	ID         int64  `json:"id"`
	LocaleCode string `json:"locale_code"`
	Slug       string `json:"slug"`
}

type PageSeoResponse struct {
	ID          int64  `json:"id"`
	LocaleCode  string `json:"locale_code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PageResponse struct {
	ID          int64              `json:"id"`
	SiteID      int64              `json:"site_id"`
	BlogID      *int64             `json:"blog_id"`
	ParentID    *int64             `json:"parent_id"`
	Type        string             `json:"type"`
	Layout      json.RawMessage    `json:"layout"`
	PublishedAt *string            `json:"published_at"`
	Seo         []PageSeoResponse  `json:"seo"`
	Slugs       []PageSlugResponse `json:"slugs"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type RouteEntry struct {
	Path   string `json:"path"`
	PageID *int64 `json:"page_id,omitempty"`
	BlogID *int64 `json:"blog_id,omitempty"`
}

type RoutesResponse struct {
	Routes map[string][]RouteEntry `json:"routes"`
}

func (h *PageHandler) Create(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	var req CreatePageRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	locale := c.QueryParam("locale")
	if locale == "" {
		return apierror.JSON(c, http.StatusBadRequest, "locale is required")
	}

	if req.Type != "page" && req.Type != "post" {
		return apierror.JSON(c, http.StatusBadRequest, "type must be 'page' or 'post'")
	}

	if req.Type == "page" && req.BlogID != nil {
		return apierror.JSON(c, http.StatusBadRequest, "blog_id is not allowed for pages")
	}

	if len(req.Slugs) == 0 {
		return apierror.JSON(c, http.StatusBadRequest, "at least one slug is required")
	}

	for i := range req.Slugs {
		if req.Slugs[i].LocaleCode == "" {
			return apierror.JSON(c, http.StatusBadRequest, "locale_code is required for each slug")
		}
		req.Slugs[i].Slug = strings.Trim(strings.TrimSpace(req.Slugs[i].Slug), "/")
		if err := validateSlug(req.Slugs[i].Slug); err != nil {
			return apierror.JSON(c, http.StatusBadRequest, err.Error())
		}
		if req.Slugs[i].Slug == "" && req.ParentID != nil {
			return apierror.JSON(c, http.StatusBadRequest, "slug cannot be empty for child pages")
		}
	}

	seen := make(map[string]bool, len(req.Slugs))
	for _, s := range req.Slugs {
		if seen[s.LocaleCode] {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("duplicate locale_code '%s' in slugs", s.LocaleCode))
		}
		seen[s.LocaleCode] = true
	}

	for _, s := range req.Seo {
		if s.LocaleCode == "" {
			return apierror.JSON(c, http.StatusBadRequest, "locale_code is required for each seo entry")
		}
		if s.Title == "" {
			return apierror.JSON(c, http.StatusBadRequest, "title is required for each seo entry")
		}
	}

	seenSeo := make(map[string]bool, len(req.Seo))
	for _, s := range req.Seo {
		if seenSeo[s.LocaleCode] {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("duplicate locale_code '%s' in seo", s.LocaleCode))
		}
		seenSeo[s.LocaleCode] = true
	}

	ctx := c.Request().Context()

	layout := req.Layout
	if layout == nil || string(layout) == "{}" || string(layout) == "null" {
		var err2 error
		layout, err2 = makeDefaultLayout(func() (int64, error) {
			return h.queries.GetNextSectionID(ctx, siteID)
		}, 4)
		if err2 != nil {
			return apierror.Internal(c, "failed to generate section id", err2)
		}
	}

	var publishedAt pgtype.Timestamptz
	if req.PublishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return apierror.JSON(c, http.StatusBadRequest, "invalid published_at format, use ISO 8601")
		}
		publishedAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	// Verify site exists
	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	// Validate blog_id for posts
	var blogID pgtype.Int8
	if req.Type == "post" {
		if req.BlogID == nil {
			return apierror.JSON(c, http.StatusBadRequest, "blog_id is required for posts")
		}
		_, err = h.queries.GetBlogByID(ctx, db.GetBlogByIDParams{ID: *req.BlogID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierror.JSON(c, http.StatusBadRequest, "blog not found in this site")
			}
			return apierror.Internal(c, "failed to get blog", err)
		}
		blogID = pgtype.Int8{Int64: *req.BlogID, Valid: true}
	}

	// Validate parent_id
	var parentID pgtype.Int8
	if req.ParentID != nil {
		parent, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: *req.ParentID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierror.JSON(c, http.StatusBadRequest, "parent page not found in this site")
			}
			return apierror.Internal(c, "failed to get parent page", err)
		}
		if parent.Type != req.Type {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("parent must be of type '%s'", req.Type))
		}
		// For posts, validate same blog
		if req.Type == "post" && parent.BlogID.Int64 != blogID.Int64 {
			return apierror.JSON(c, http.StatusBadRequest, "parent post must belong to the same blog")
		}
		parentID = pgtype.Int8{Int64: *req.ParentID, Valid: true}
	}

	// Validate locale IDs belong to site
	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get site locales", err)
	}
	localeByCode := make(map[string]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeByCode[l.Code] = l
	}
	for _, s := range req.Slugs {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}
	for _, s := range req.Seo {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}

	if _, ok := localeByCode[locale]; !ok {
		return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale '%s' does not belong to this site", locale))
	}

	layout, err = wrapLayoutLocales(layout, locale)
	if err != nil {
		return apierror.Internal(c, "failed to wrap layout locales", err)
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return apierror.Internal(c, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	page, err := qtx.CreatePage(ctx, db.CreatePageParams{
		SiteID:      siteID,
		BlogID:      blogID,
		ParentID:    parentID,
		Type:        req.Type,
		Layout:      layout,
		PublishedAt: publishedAt,
	})
	if err != nil {
		return apierror.Internal(c, "failed to create page", err)
	}

	var slugs []PageSlugResponse
	for _, s := range req.Slugs {
		loc := localeByCode[s.LocaleCode]
		slug, err := qtx.CreatePageSlug(ctx, db.CreatePageSlugParams{
			PageID:   page.ID,
			LocaleID: loc.ID,
			Slug:     s.Slug,
		})
		if err != nil {
			return apierror.Internal(c, "failed to create page slug", err)
		}
		slugs = append(slugs, PageSlugResponse{
			ID:         slug.ID,
			LocaleCode: loc.Code,
			Slug:       slug.Slug,
		})
	}

	var seo []PageSeoResponse
	for _, s := range req.Seo {
		loc := localeByCode[s.LocaleCode]
		var desc pgtype.Text
		if s.Description != "" {
			desc = pgtype.Text{String: s.Description, Valid: true}
		}
		row, err := qtx.CreatePageSeo(ctx, db.CreatePageSeoParams{
			PageID:      page.ID,
			LocaleID:    loc.ID,
			Title:       s.Title,
			Description: desc,
		})
		if err != nil {
			return apierror.Internal(c, "failed to create page seo", err)
		}
		seo = append(seo, PageSeoResponse{
			ID:          row.ID,
			LocaleCode:  loc.Code,
			Title:       row.Title,
			Description: row.Description.String,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return apierror.Internal(c, "failed to commit transaction", err)
	}

	resp := toPageResponse(page, slugs, seo)
	resp.Layout, err = resolveLayoutLocales(resp.Layout, locale)
	if err != nil {
		return apierror.Internal(c, "failed to resolve layout locales", err)
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *PageHandler) GetByID(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	pageID, err := parseID(c.Param("pageId"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	locale := c.QueryParam("locale")
	if locale == "" {
		return apierror.JSON(c, http.StatusBadRequest, "locale is required")
	}

	ctx := c.Request().Context()

	page, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "page not found")
		}
		return apierror.Internal(c, "failed to get page", err)
	}

	slugs, err := h.queries.GetPageSlugsByPageID(ctx, pageID)
	if err != nil {
		return apierror.Internal(c, "failed to get page slugs", err)
	}

	seoRows, err := h.queries.GetPageSeoByPageID(ctx, pageID)
	if err != nil {
		return apierror.Internal(c, "failed to get page seo", err)
	}

	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get site locales", err)
	}
	localeMap := make(map[int64]db.SiteLocale, len(siteLocales))
	validLocale := false
	for _, l := range siteLocales {
		localeMap[l.ID] = l
		if l.Code == locale {
			validLocale = true
		}
	}
	if !validLocale {
		return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale '%s' does not belong to this site", locale))
	}

	resp := toPageResponse(page, toSlugResponses(slugs, localeMap), toSeoResponses(seoRows, localeMap))
	resp.Layout, err = resolveLayoutLocales(resp.Layout, locale)
	if err != nil {
		return apierror.Internal(c, "failed to resolve layout locales", err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *PageHandler) List(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	// Verify site exists
	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

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
		"id":           "id",
		"parent_id":    "parent_id",
		"type":         "type",
		"blog_id":      "blog_id",
		"published_at": "published_at",
	})
	if err := builder.Parse(c.Request().URL.Query()); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	result := builder.Build(cursor, 2)

	whereClause := " WHERE site_id = $1"
	args := []any{siteID}
	if result.WhereClause != "" {
		whereClause += " AND " + result.WhereClause
		args = append(args, result.Args...)
	}

	countSQL := "SELECT COUNT(*) FROM pages" + whereClause
	var total int64
	if err := h.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return apierror.Internal(c, "failed to count pages", err)
	}

	listSQL := "SELECT id, site_id, blog_id, parent_id, type, layout, published_at, created_at, updated_at FROM pages" + whereClause + " ORDER BY id ASC"

	paginated := limit != nil
	var listArgs []any
	if paginated {
		nextParam := len(args) + 1
		listSQL += fmt.Sprintf(" LIMIT $%d", nextParam)
		listArgs = append(append([]any{}, args...), int64(*limit))
	} else {
		listArgs = args
	}

	rows, err := h.pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return apierror.Internal(c, "failed to list pages", err)
	}
	defer rows.Close()

	var pages []db.Page
	for rows.Next() {
		var p db.Page
		if err := rows.Scan(&p.ID, &p.SiteID, &p.BlogID, &p.ParentID, &p.Type, &p.Layout, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return apierror.Internal(c, "failed to scan page", err)
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return apierror.Internal(c, "failed to list pages", err)
	}

	pageIDs := make([]int64, len(pages))
	for i, p := range pages {
		pageIDs[i] = p.ID
	}

	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get site locales", err)
	}
	localeMap := make(map[int64]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeMap[l.ID] = l
	}

	var slugsByPage map[int64][]PageSlugResponse
	var seoByPage map[int64][]PageSeoResponse
	if len(pageIDs) > 0 {
		allSlugs, err := h.queries.GetPageSlugsByPageIDs(ctx, pageIDs)
		if err != nil {
			return apierror.Internal(c, "failed to get page slugs", err)
		}
		slugsByPage = make(map[int64][]PageSlugResponse, len(pageIDs))
		for _, s := range allSlugs {
			loc := localeMap[s.LocaleID]
			slugsByPage[s.PageID] = append(slugsByPage[s.PageID], PageSlugResponse{
				ID:         s.ID,
				LocaleCode: loc.Code,
				Slug:       s.Slug,
			})
		}

		allSeo, err := h.queries.GetPageSeoByPageIDs(ctx, pageIDs)
		if err != nil {
			return apierror.Internal(c, "failed to get page seo", err)
		}
		seoByPage = make(map[int64][]PageSeoResponse, len(pageIDs))
		for _, s := range allSeo {
			loc := localeMap[s.LocaleID]
			seoByPage[s.PageID] = append(seoByPage[s.PageID], PageSeoResponse{
				ID:          s.ID,
				LocaleCode:  loc.Code,
				Title:       s.Title,
				Description: s.Description.String,
			})
		}
	}

	data := make([]PageResponse, 0, len(pages))
	for _, p := range pages {
		slugs := slugsByPage[p.ID]
		if slugs == nil {
			slugs = []PageSlugResponse{}
		}
		seo := seoByPage[p.ID]
		if seo == nil {
			seo = []PageSeoResponse{}
		}
		data = append(data, toPageResponse(p, slugs, seo))
	}

	var nextCursor *int64
	var hasMore bool
	if paginated {
		hasMore = len(pages) == int(*limit)
		if hasMore && len(pages) > 0 {
			lastID := pages[len(pages)-1].ID
			nextCursor = &lastID
		}
	}

	return c.JSON(http.StatusOK, PaginatedResponse[PageResponse]{
		Data:       data,
		NextCursor: nextCursor,
		Total:      total,
		HasMore:    hasMore,
	})
}

func (h *PageHandler) Update(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	pageID, err := parseID(c.Param("pageId"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	var req UpdatePageRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	locale := c.QueryParam("locale")
	if locale == "" {
		return apierror.JSON(c, http.StatusBadRequest, "locale is required")
	}

	if len(req.Slugs) == 0 {
		return apierror.JSON(c, http.StatusBadRequest, "at least one slug is required")
	}

	for i := range req.Slugs {
		if req.Slugs[i].LocaleCode == "" {
			return apierror.JSON(c, http.StatusBadRequest, "locale_code is required for each slug")
		}
		req.Slugs[i].Slug = strings.Trim(strings.TrimSpace(req.Slugs[i].Slug), "/")
		if err := validateSlug(req.Slugs[i].Slug); err != nil {
			return apierror.JSON(c, http.StatusBadRequest, err.Error())
		}
	}

	seen := make(map[string]bool, len(req.Slugs))
	for _, s := range req.Slugs {
		if seen[s.LocaleCode] {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("duplicate locale_code '%s' in slugs", s.LocaleCode))
		}
		seen[s.LocaleCode] = true
	}

	for _, s := range req.Seo {
		if s.LocaleCode == "" {
			return apierror.JSON(c, http.StatusBadRequest, "locale_code is required for each seo entry")
		}
		if s.Title == "" {
			return apierror.JSON(c, http.StatusBadRequest, "title is required for each seo entry")
		}
	}

	seenSeo := make(map[string]bool, len(req.Seo))
	for _, s := range req.Seo {
		if seenSeo[s.LocaleCode] {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("duplicate locale_code '%s' in seo", s.LocaleCode))
		}
		seenSeo[s.LocaleCode] = true
	}

	ctx := c.Request().Context()

	// Get existing page to know its type, blog_id and layout
	existing, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "page not found")
		}
		return apierror.Internal(c, "failed to get page", err)
	}

	var layout json.RawMessage
	if req.Layout != nil {
		layout, err = mergeLayoutLocales(req.Layout, existing.Layout, locale)
		if err != nil {
			return apierror.Internal(c, "failed to merge layout locales", err)
		}
	} else {
		layout = existing.Layout
	}

	var publishedAt pgtype.Timestamptz
	if req.PublishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return apierror.JSON(c, http.StatusBadRequest, "invalid published_at format, use ISO 8601")
		}
		publishedAt = pgtype.Timestamptz{Time: t, Valid: true}
	} else {
		publishedAt = existing.PublishedAt
	}

	// Validate locales
	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get site locales", err)
	}
	localeByCode := make(map[string]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeByCode[l.Code] = l
	}
	for _, s := range req.Slugs {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}
	for _, s := range req.Seo {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}

	if _, ok := localeByCode[locale]; !ok {
		return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("locale '%s' does not belong to this site", locale))
	}

	// Validate parent_id
	var parentID pgtype.Int8
	if req.ParentID != nil {
		if *req.ParentID == pageID {
			return apierror.JSON(c, http.StatusBadRequest, "page cannot be its own parent")
		}
		parent, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: *req.ParentID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierror.JSON(c, http.StatusBadRequest, "parent page not found in this site")
			}
			return apierror.Internal(c, "failed to get parent page", err)
		}
		if parent.Type != existing.Type {
			return apierror.JSON(c, http.StatusBadRequest, fmt.Sprintf("parent must be of type '%s'", existing.Type))
		}
		if existing.Type == "post" && parent.BlogID != existing.BlogID {
			return apierror.JSON(c, http.StatusBadRequest, "parent post must belong to the same blog")
		}
		parentID = pgtype.Int8{Int64: *req.ParentID, Valid: true}
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return apierror.Internal(c, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	page, err := qtx.UpdatePage(ctx, db.UpdatePageParams{
		ParentID:    parentID,
		Layout:      layout,
		PublishedAt: publishedAt,
		ID:          pageID,
		SiteID:      siteID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "page not found")
		}
		return apierror.Internal(c, "failed to update page", err)
	}

	// Replace slugs
	if err := qtx.DeletePageSlugsByPageID(ctx, pageID); err != nil {
		return apierror.Internal(c, "failed to delete old slugs", err)
	}

	var slugs []PageSlugResponse
	for _, s := range req.Slugs {
		loc := localeByCode[s.LocaleCode]
		slug, err := qtx.CreatePageSlug(ctx, db.CreatePageSlugParams{
			PageID:   pageID,
			LocaleID: loc.ID,
			Slug:     s.Slug,
		})
		if err != nil {
			return apierror.Internal(c, "failed to create page slug", err)
		}
		slugs = append(slugs, PageSlugResponse{
			ID:         slug.ID,
			LocaleCode: loc.Code,
			Slug:       slug.Slug,
		})
	}

	// Replace seo
	if err := qtx.DeletePageSeoByPageID(ctx, pageID); err != nil {
		return apierror.Internal(c, "failed to delete old seo", err)
	}

	var seo []PageSeoResponse
	for _, s := range req.Seo {
		loc := localeByCode[s.LocaleCode]
		var desc pgtype.Text
		if s.Description != "" {
			desc = pgtype.Text{String: s.Description, Valid: true}
		}
		row, err := qtx.CreatePageSeo(ctx, db.CreatePageSeoParams{
			PageID:      pageID,
			LocaleID:    loc.ID,
			Title:       s.Title,
			Description: desc,
		})
		if err != nil {
			return apierror.Internal(c, "failed to create page seo", err)
		}
		seo = append(seo, PageSeoResponse{
			ID:          row.ID,
			LocaleCode:  loc.Code,
			Title:       row.Title,
			Description: row.Description.String,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return apierror.Internal(c, "failed to commit transaction", err)
	}

	resp := toPageResponse(page, slugs, seo)
	resp.Layout, err = resolveLayoutLocales(resp.Layout, locale)
	if err != nil {
		return apierror.Internal(c, "failed to resolve layout locales", err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *PageHandler) Delete(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	pageID, err := parseID(c.Param("pageId"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	_, err = h.queries.DeletePage(c.Request().Context(), db.DeletePageParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "page not found")
		}
		return apierror.Internal(c, "failed to delete page", err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *PageHandler) Routes(c *echo.Context) error {
	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	routes, err := buildRoutesMap(ctx, h.queries, siteID)
	if err != nil {
		return apierror.Internal(c, err.Error(), nil)
	}

	return c.JSON(http.StatusOK, RoutesResponse{Routes: routes})
}

func buildRoutePath(localeCode string, isDefault bool, slugPath string) string {
	if slugPath == "" {
		if isDefault {
			return "/"
		}
		return "/" + localeCode
	}
	if isDefault {
		return "/" + slugPath
	}
	return "/" + localeCode + "/" + slugPath
}

func toPageResponse(p db.Page, slugs []PageSlugResponse, seo []PageSeoResponse) PageResponse {
	var blogID *int64
	if p.BlogID.Valid {
		blogID = &p.BlogID.Int64
	}

	var parentID *int64
	if p.ParentID.Valid {
		parentID = &p.ParentID.Int64
	}

	var publishedAt *string
	if p.PublishedAt.Valid {
		s := p.PublishedAt.Time.Format(time.RFC3339)
		publishedAt = &s
	}

	return PageResponse{
		ID:          p.ID,
		SiteID:      p.SiteID,
		BlogID:      blogID,
		ParentID:    parentID,
		Type:        p.Type,
		Layout:      p.Layout,
		PublishedAt: publishedAt,
		Seo:         seo,
		Slugs:       slugs,
		CreatedAt:   p.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func toSlugResponses(slugs []db.PageSlug, localeMap map[int64]db.SiteLocale) []PageSlugResponse {
	result := make([]PageSlugResponse, 0, len(slugs))
	for _, s := range slugs {
		loc := localeMap[s.LocaleID]
		result = append(result, PageSlugResponse{
			ID:         s.ID,
			LocaleCode: loc.Code,
			Slug:       s.Slug,
		})
	}
	return result
}

func toSeoResponses(seoRows []db.GetPageSeoByPageIDRow, localeMap map[int64]db.SiteLocale) []PageSeoResponse {
	result := make([]PageSeoResponse, 0, len(seoRows))
	for _, s := range seoRows {
		loc := localeMap[s.LocaleID]
		result = append(result, PageSeoResponse{
			ID:          s.ID,
			LocaleCode:  loc.Code,
			Title:       s.Title,
			Description: s.Description.String,
		})
	}
	return result
}
