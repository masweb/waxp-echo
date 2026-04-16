package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
)

type LocaleResponse struct {
	Code      string `json:"code"`
	IsDefault bool   `json:"is_default"`
}

type LocaleHandler struct {
	queries *db.Queries
}

func NewLocaleHandler(queries *db.Queries) *LocaleHandler {
	return &LocaleHandler{queries: queries}
}

type AddLocaleRequest struct {
	Code      string `json:"code" validate:"required,min=1,max=10"`
	IsDefault bool   `json:"is_default"`
}

func (h *LocaleHandler) Add(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid site id")
	}

	var req AddLocaleRequest
	if err := c.Bind(&req); err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Code == "" {
		return apierror.JSON(c, http.StatusBadRequest, "code is required")
	}

	_, err = h.queries.GetSiteByID(c.Request().Context(), siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	if req.IsDefault {
		err = h.queries.UnsetDefaultLocales(c.Request().Context(), siteID)
		if err != nil {
			return apierror.Internal(c, "failed to update default locale", err)
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
			return apierror.JSON(c, http.StatusConflict, "locale code already exists for this site")
		}
		return apierror.Internal(c, "failed to create locale", err)
	}

	return c.JSON(http.StatusCreated, LocaleResponse{
		Code:      locale.Code,
		IsDefault: locale.IsDefault,
	})
}

func (h *LocaleHandler) Remove(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid site id")
	}

	localeCode := c.Param("localeCode")

	_, err = h.queries.GetSiteByID(c.Request().Context(), siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	_, err = h.queries.GetSiteLocaleByCodeAndSite(c.Request().Context(), db.GetSiteLocaleByCodeAndSiteParams{
		Code:   localeCode,
		SiteID: siteID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "locale not found")
		}
		return apierror.Internal(c, "failed to get locale", err)
	}

	err = h.queries.DeleteSiteLocaleByCode(c.Request().Context(), db.DeleteSiteLocaleByCodeParams{
		Code:   localeCode,
		SiteID: siteID,
	})
	if err != nil {
		return apierror.Internal(c, "failed to delete locale", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}
