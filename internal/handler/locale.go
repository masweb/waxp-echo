package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/db"
)

type LocaleResponse struct {
	Code      string `json:"code"`
	IsDefault bool   `json:"is_default"`
}

type LocaleHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewLocaleHandler(queries *db.Queries, pool *pgxpool.Pool) *LocaleHandler {
	return &LocaleHandler{queries: queries, pool: pool}
}

type AddLocaleRequest struct {
	Code      string `json:"code" validate:"required,min=1,max=10"`
	IsDefault bool   `json:"is_default"`
}

func (h *LocaleHandler) Add(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	var req AddLocaleRequest
	if err := c.Bind(&req); err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Code == "" {
		return ErrorJSON(c, http.StatusBadRequest, "code is required")
	}

	_, err = h.queries.GetSiteByID(c.Request().Context(), siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

	if req.IsDefault {
		err = h.unsetDefaults(c.Request().Context(), siteID)
		if err != nil {
			return InternalError(c, "failed to update default locale", err)
		}
	}

	locale, err := h.queries.CreateSiteLocale(c.Request().Context(), db.CreateSiteLocaleParams{
		SiteID:    siteID,
		Code:      req.Code,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrorJSON(c, http.StatusConflict, "locale code already exists for this site")
		}
		return InternalError(c, "failed to create locale", err)
	}

	return c.JSON(http.StatusCreated, LocaleResponse{
		Code:      locale.Code,
		IsDefault: locale.IsDefault,
	})
}

func (h *LocaleHandler) Remove(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorJSON(c, http.StatusBadRequest, "invalid site id")
	}

	localeCode := c.Param("localeCode")

	_, err = h.queries.GetSiteByID(c.Request().Context(), siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "site not found")
		}
		return InternalError(c, "failed to get site", err)
	}

	_, err = h.queries.GetSiteLocaleByCodeAndSite(c.Request().Context(), db.GetSiteLocaleByCodeAndSiteParams{
		Code:   localeCode,
		SiteID: siteID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrorJSON(c, http.StatusNotFound, "locale not found")
		}
		return InternalError(c, "failed to get locale", err)
	}

	err = h.queries.DeleteSiteLocaleByCode(c.Request().Context(), db.DeleteSiteLocaleByCodeParams{
		Code:   localeCode,
		SiteID: siteID,
	})
	if err != nil {
		return InternalError(c, "failed to delete locale", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

func (h *LocaleHandler) unsetDefaults(ctx context.Context, siteID int64) error {
	_, err := h.pool.Exec(ctx, "UPDATE site_locales SET is_default = false WHERE site_id = $1 AND is_default = true", siteID)
	return err
}
