package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
	"waxp/echo/internal/i18n"
	"waxp/echo/internal/render"
)

type PublicHandler struct {
	queries   *db.Queries
	pool      *pgxpool.Pool
	mediaBase string
}

func NewPublicHandler(queries *db.Queries, pool *pgxpool.Pool, mediaBase string) *PublicHandler {
	return &PublicHandler{queries: queries, pool: pool, mediaBase: mediaBase}
}

func (h *PublicHandler) ServePage(c *echo.Context) error {
	ctx := c.Request().Context()

	site, err := h.queries.GetLiveSite(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.String(http.StatusNotFound, "no live site configured")
		}
		return apierror.Internal(c, "failed to get live site", err)
	}

	locales, err := h.queries.ListSiteLocales(ctx, site.ID)
	if err != nil {
		return apierror.Internal(c, "failed to get site locales", err)
	}

	if len(locales) == 0 {
		return c.String(http.StatusNotFound, "site has no locales")
	}

	path := strings.TrimPrefix(c.Request().URL.Path, "/")
	path = strings.TrimSuffix(path, "/")

	localeCode := ""
	slug := ""

	if path == "" {
		for _, l := range locales {
			if l.IsDefault {
				localeCode = l.Code
				break
			}
		}
	} else {
		parts := strings.SplitN(path, "/", 2)
		first := parts[0]

		isLocale := false
		for _, l := range locales {
			if l.Code == first {
				isLocale = true
				localeCode = l.Code
				break
			}
		}

		if isLocale {
			if len(parts) > 1 {
				slug = parts[1]
			}
		} else {
			for _, l := range locales {
				if l.IsDefault {
					localeCode = l.Code
					break
				}
			}
			slug = path
		}
	}

	if localeCode == "" {
		localeCode = locales[0].Code
	}

	slugRow, err := h.queries.GetPublishedPageSlug(ctx, db.GetPublishedPageSlugParams{
		SiteID: site.ID,
		Slug:   slug,
		Code:   localeCode,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.String(http.StatusNotFound, "page not found")
		}
		return apierror.Internal(c, "failed to lookup page", err)
	}

	renderRow, err := h.queries.GetPageRenderByPageAndLocale(ctx, db.GetPageRenderByPageAndLocaleParams{
		PageID:  slugRow.PageID,
		LocaleID: slugRow.LocaleID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.String(http.StatusNotFound, "page not rendered yet")
		}
		return apierror.Internal(c, "failed to get page render", err)
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return c.String(http.StatusOK, renderRow.Html)
}

func (h *PublicHandler) RegenerateAllPages(c *echo.Context) error {
	ctx := c.Request().Context()

	siteID, err := parseID(c.Param("id"))
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, err.Error())
	}

	site, err := h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	locales, err := h.queries.ListSiteLocales(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get locales", err)
	}

	pageIDs, err := h.queries.GetAllPublishedPageIDs(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get pages", err)
	}

	count := 0
	for _, pid := range pageIDs {
		for _, loc := range locales {
			_, err := renderPageForLocale(ctx, h.queries, site, pid, loc, h.mediaBase)
			if err != nil {
				return apierror.Internal(c, "failed to render page", err)
			}
			count++
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"regenerated": count,
	})
}

func renderPageForLocale(ctx context.Context, queries *db.Queries, site db.Site, pageID int64, locale db.SiteLocale, mediaBase string) (db.PageRender, error) {
	page, err := queries.GetPageByID(ctx, db.GetPageByIDParams{ID: pageID, SiteID: site.ID})
	if err != nil {
		return db.PageRender{}, err
	}

	resolvedLayout, err := i18n.Resolve(page.Layout, locale.Code)
	if err != nil {
		return db.PageRender{}, err
	}

	resolvedOptions, err := i18n.Resolve(site.Options, locale.Code)
	if err != nil {
		return db.PageRender{}, err
	}

	seoRows, err := queries.GetPageSeoByPageID(ctx, pageID)
	if err != nil {
		return db.PageRender{}, err
	}

	var seo *render.SEOData
	for _, s := range seoRows {
		if s.LocaleID == locale.ID {
			seo = &render.SEOData{
				Title:       s.Title,
				Description: s.Description.String,
			}
			break
		}
	}

	slugs, err := queries.GetPageSlugsByPageID(ctx, pageID)
	if err != nil {
		return db.PageRender{}, err
	}

	allLocales, err := queries.ListSiteLocales(ctx, site.ID)
	if err != nil {
		return db.PageRender{}, err
	}

	localeMap := make(map[int64]db.SiteLocale, len(allLocales))
	for _, l := range allLocales {
		localeMap[l.ID] = l
	}

	var pageSlugs []render.SlugInfo
	for _, s := range slugs {
		loc := localeMap[s.LocaleID]
		pageSlugs = append(pageSlugs, render.SlugInfo{
			LocaleCode: loc.Code,
			Slug:       s.Slug,
			IsDefault:  loc.IsDefault,
		})
	}

	var localeInfos []render.LocaleInfo
	for _, l := range allLocales {
		localeInfos = append(localeInfos, render.LocaleInfo{
			Code:      l.Code,
			IsDefault: l.IsDefault,
		})
	}

	html, err := render.Render(render.RenderInput{
		LayoutJSON:  resolvedLayout,
		OptionsJSON: resolvedOptions,
		SEO:         seo,
		Locale:      locale.Code,
		Locales:     localeInfos,
		PageSlugs:   pageSlugs,
		Domain:      site.Domain,
		MediaBase:   mediaBase,
	})
	if err != nil {
		return db.PageRender{}, err
	}

	renderRow, err := queries.UpsertPageRender(ctx, db.UpsertPageRenderParams{
		PageID:   pageID,
		LocaleID: locale.ID,
		Html:     html,
	})
	if err != nil {
		return db.PageRender{}, err
	}

	return renderRow, nil
}
