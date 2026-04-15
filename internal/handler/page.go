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
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	var req CreatePageRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Type != "page" && req.Type != "post" {
		return ErrorJSON(c, http.StatusBadRequest, "type must be 'page' or 'post'")
	}

	if len(req.Slugs) == 0 {
		return ErrorJSON(c, http.StatusBadRequest, "at least one slug is required")
	}

	for i := range req.Slugs {
		req.Slugs[i].Slug = strings.Trim(strings.TrimSpace(req.Slugs[i].Slug), "/")
		if req.Slugs[i].Slug == "" && req.ParentID != nil {
			return ErrorJSON(c, http.StatusBadRequest, "slug cannot be empty for child pages")
		}
	}

	for _, s := range req.Seo {
		if s.LocaleCode == "" {
			return ErrorJSON(c, http.StatusBadRequest, "locale_code is required for each seo entry")
		}
		if s.Title == "" {
			return ErrorJSON(c, http.StatusBadRequest, "title is required for each seo entry")
		}
	}

	ctx := c.Request().Context()

	layout := req.Layout
	if layout == nil || string(layout) == "{}" || string(layout) == "null" {
		sectionIDs := make([]int64, 4)
		for i := range sectionIDs {
			id, err := h.queries.GetNextSectionID(ctx, siteID)
			if err != nil {
				return InternalError(c, "failed to generate section id", err)
			}
			sectionIDs[i] = id
		}
		defaultLayout := []map[string]interface{}{
			{"id": sectionIDs[0], "mobile": map[string]int{"cols": 8, "rows": 12}, "tablet": map[string]int{"cols": 12, "rows": 16}, "desktop": map[string]int{"cols": 24, "rows": 20}, "blocks": []interface{}{}},
			{"id": sectionIDs[1], "mobile": map[string]int{"cols": 8, "rows": 12}, "tablet": map[string]int{"cols": 12, "rows": 16}, "desktop": map[string]int{"cols": 24, "rows": 20}, "blocks": []interface{}{}},
			{"id": sectionIDs[2], "mobile": map[string]int{"cols": 8, "rows": 12}, "tablet": map[string]int{"cols": 12, "rows": 16}, "desktop": map[string]int{"cols": 24, "rows": 20}, "blocks": []interface{}{}},
			{"id": sectionIDs[3], "mobile": map[string]int{"cols": 8, "rows": 12}, "tablet": map[string]int{"cols": 12, "rows": 16}, "desktop": map[string]int{"cols": 24, "rows": 20}, "blocks": []interface{}{}},
		}
		layout, _ = json.Marshal(defaultLayout)
	}

	var publishedAt pgtype.Timestamptz
	if req.PublishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return ErrorJSON(c, http.StatusBadRequest, "invalid published_at format, use ISO 8601")
		}
		publishedAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	// Verify site exists
	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

	// Validate blog_id for posts
	var blogID pgtype.Int8
	if req.Type == "post" {
		if req.BlogID == nil {
			return ErrorJSON(c, http.StatusBadRequest, "blog_id is required for posts")
		}
		_, err = h.queries.GetBlogByID(ctx, db.GetBlogByIDParams{ID: *req.BlogID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrorJSON(c, http.StatusBadRequest, "blog not found in this site")
			}
			return InternalError(c, "failed to get blog", err)
		}
		blogID = pgtype.Int8{Int64: *req.BlogID, Valid: true}
	}

	// Validate parent_id
	var parentID pgtype.Int8
	if req.ParentID != nil {
		parent, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: *req.ParentID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrorJSON(c, http.StatusBadRequest, "parent page not found in this site")
			}
			return InternalError(c, "failed to get parent page", err)
		}
		if parent.Type != req.Type {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("parent must be of type '%s'", req.Type))
		}
		// For posts, validate same blog
		if req.Type == "post" && parent.BlogID.Int64 != blogID.Int64 {
			return ErrorJSON(c, http.StatusBadRequest, "parent post must belong to the same blog")
		}
		parentID = pgtype.Int8{Int64: *req.ParentID, Valid: true}
	}

	// Validate locale IDs belong to site
	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return InternalError(c, "failed to get site locales", err)
	}
	localeByCode := make(map[string]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeByCode[l.Code] = l
	}
	for _, s := range req.Slugs {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}
	for _, s := range req.Seo {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return InternalError(c, "failed to begin transaction", err)
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
		return InternalError(c, "failed to create page", err)
	}

	localeByID := make(map[int64]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeByID[l.ID] = l
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
			return InternalError(c, "failed to create page slug", err)
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
			return InternalError(c, "failed to create page seo", err)
		}
		seo = append(seo, PageSeoResponse{
			ID:          row.ID,
			LocaleCode:  loc.Code,
			Title:       row.Title,
			Description: row.Description.String,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return InternalError(c, "failed to commit transaction", err)
	}

	return c.JSON(http.StatusCreated, toPageResponse(page, slugs, seo))
}

func (h *PageHandler) GetByID(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	pageID, err := strconv.ParseInt(c.Param("pageId"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid page id")
	}

	ctx := c.Request().Context()

	page, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "page not found")
		}
		return InternalError(c, "failed to get page", err)
	}

	slugs, err := h.queries.GetPageSlugsByPageID(ctx, pageID)
	if err != nil {
		return InternalError(c, "failed to get page slugs", err)
	}

	seoRows, err := h.queries.GetPageSeoByPageID(ctx, pageID)
	if err != nil {
		return InternalError(c, "failed to get page seo", err)
	}

	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return InternalError(c, "failed to get site locales", err)
	}
	localeMap := make(map[int64]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeMap[l.ID] = l
	}

	return c.JSON(http.StatusOK, toPageResponse(page, toSlugResponses(slugs, localeMap), toSeoResponses(seoRows, localeMap)))
}

func (h *PageHandler) List(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	ctx := c.Request().Context()

	// Verify site exists
	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

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
		"id":           "id",
		"parent_id":    "parent_id",
		"type":         "type",
		"blog_id":      "blog_id",
		"published_at": "published_at",
	})
	if err := builder.Parse(c.Request().URL.Query()); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, err.Error())
	}

	result := builder.Build(cursor)

	var whereClause string
	var args []any
	if result.WhereClause != "" {
		filterClause := shiftParamIndices(result.WhereClause[len(" WHERE "):], 1)
		whereClause = " WHERE site_id = $1 AND " + filterClause
		args = append([]any{siteID}, result.Args...)
	} else {
		whereClause = " WHERE site_id = $1"
		args = []any{siteID}
	}

	countSQL := "SELECT COUNT(*) FROM pages" + whereClause
	var total int64
	if err := h.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return InternalError(c, "failed to count pages", err)
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
		return InternalError(c, "failed to list pages", err)
	}
	defer rows.Close()

	var pages []db.Page
	for rows.Next() {
		var p db.Page
		if err := rows.Scan(&p.ID, &p.SiteID, &p.BlogID, &p.ParentID, &p.Type, &p.Layout, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return InternalError(c, "failed to scan page", err)
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return InternalError(c, "failed to list pages", err)
	}

	data := make([]PageResponse, 0, len(pages))

	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return InternalError(c, "failed to get site locales", err)
	}
	localeMap := make(map[int64]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeMap[l.ID] = l
	}

	for _, p := range pages {
		slugs, err := h.queries.GetPageSlugsByPageID(ctx, p.ID)
		if err != nil {
			return InternalError(c, "failed to get page slugs", err)
		}
		seoRows, err := h.queries.GetPageSeoByPageID(ctx, p.ID)
		if err != nil {
			return InternalError(c, "failed to get page seo", err)
		}
		data = append(data, toPageResponse(p, toSlugResponses(slugs, localeMap), toSeoResponses(seoRows, localeMap)))
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
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	pageID, err := strconv.ParseInt(c.Param("pageId"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid page id")
	}

	var req UpdatePageRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if len(req.Slugs) == 0 {
		return ErrorJSON(c, http.StatusBadRequest, "at least one slug is required")
	}

	for i := range req.Slugs {
		req.Slugs[i].Slug = strings.Trim(strings.TrimSpace(req.Slugs[i].Slug), "/")
	}

	for _, s := range req.Seo {
		if s.LocaleCode == "" {
			return ErrorJSON(c, http.StatusBadRequest, "locale_code is required for each seo entry")
		}
		if s.Title == "" {
			return ErrorJSON(c, http.StatusBadRequest, "title is required for each seo entry")
		}
	}

	ctx := c.Request().Context()

	// Get existing page to know its type, blog_id and layout
	existing, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "page not found")
		}
		return InternalError(c, "failed to get page", err)
	}

	layout := req.Layout
	if layout == nil {
		layout = existing.Layout
	}

	var publishedAt pgtype.Timestamptz
	if req.PublishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return ErrorJSON(c, http.StatusBadRequest, "invalid published_at format, use ISO 8601")
		}
		publishedAt = pgtype.Timestamptz{Time: t, Valid: true}
	} else {
		publishedAt = existing.PublishedAt
	}

	// Validate locales
	siteLocales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return InternalError(c, "failed to get site locales", err)
	}
	localeByCode := make(map[string]db.SiteLocale, len(siteLocales))
	for _, l := range siteLocales {
		localeByCode[l.Code] = l
	}
	for _, s := range req.Slugs {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}
	for _, s := range req.Seo {
		if _, ok := localeByCode[s.LocaleCode]; !ok {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("locale_code '%s' does not belong to this site", s.LocaleCode))
		}
	}

	// Validate parent_id
	var parentID pgtype.Int8
	if req.ParentID != nil {
		if *req.ParentID == pageID {
			return ErrorJSON(c, http.StatusBadRequest, "page cannot be its own parent")
		}
		parent, err := h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: *req.ParentID, SiteID: siteID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrorJSON(c, http.StatusBadRequest, "parent page not found in this site")
			}
			return InternalError(c, "failed to get parent page", err)
		}
		if parent.Type != existing.Type {
			return ErrorJSON(c, http.StatusBadRequest, fmt.Sprintf("parent must be of type '%s'", existing.Type))
		}
		if existing.Type == "post" && parent.BlogID != existing.BlogID {
			return ErrorJSON(c, http.StatusBadRequest, "parent post must belong to the same blog")
		}
		parentID = pgtype.Int8{Int64: *req.ParentID, Valid: true}
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return InternalError(c, "failed to begin transaction", err)
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
			return ErrorJSON(c, http.StatusNotFound, "page not found")
		}
		return InternalError(c, "failed to update page", err)
	}

	// Replace slugs
	if err := qtx.DeletePageSlugsByPageID(ctx, pageID); err != nil {
		return InternalError(c, "failed to delete old slugs", err)
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
			return InternalError(c, "failed to create page slug", err)
		}
		slugs = append(slugs, PageSlugResponse{
			ID:         slug.ID,
			LocaleCode: loc.Code,
			Slug:       slug.Slug,
		})
	}

	// Replace seo
	if err := qtx.DeletePageSeoByPageID(ctx, pageID); err != nil {
		return InternalError(c, "failed to delete old seo", err)
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
			return InternalError(c, "failed to create page seo", err)
		}
		seo = append(seo, PageSeoResponse{
			ID:          row.ID,
			LocaleCode:  loc.Code,
			Title:       row.Title,
			Description: row.Description.String,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return InternalError(c, "failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, toPageResponse(page, slugs, seo))
}

func (h *PageHandler) Delete(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	pageID, err := strconv.ParseInt(c.Param("pageId"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid page id")
	}

	ctx := c.Request().Context()

	_, err = h.queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: siteID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "page not found")
		}
		return InternalError(c, "failed to get page", err)
	}

	if err := h.queries.DeletePage(ctx, db.DeletePageParams{ID: pageID, SiteID: siteID}); err != nil {
		return InternalError(c, "failed to delete page", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

func (h *PageHandler) Routes(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	ctx := c.Request().Context()

	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

	routes, err := buildRoutesMap(ctx, h.queries, siteID)
	if err != nil {
		return InternalError(c, err.Error(), nil)
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

// shiftParamIndices shifts all $N parameter references in a WHERE clause by offset.
// e.g. "id > $1 AND name = $2" with offset 1 becomes "id > $2 AND name = $3"
func shiftParamIndices(clause string, offset int) string {
	var result []byte
	i := 0
	for i < len(clause) {
		if clause[i] == '$' && i+1 < len(clause) && clause[i+1] >= '0' && clause[i+1] <= '9' {
			result = append(result, '$')
			i++
			num := 0
			for i < len(clause) && clause[i] >= '0' && clause[i] <= '9' {
				num = num*10 + int(clause[i]-'0')
				i++
			}
			result = append(result, []byte(strconv.Itoa(num+offset))...)
		} else {
			result = append(result, clause[i])
			i++
		}
	}
	return string(result)
}
