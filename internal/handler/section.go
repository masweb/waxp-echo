package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
	"waxp/echo/internal/db"
)

type SectionHandler struct {
	queries *db.Queries
}

func NewSectionHandler(queries *db.Queries) *SectionHandler {
	return &SectionHandler{queries: queries}
}

type NextSectionIDResponse struct {
	ID int64 `json:"id"`
}

func (h *SectionHandler) GetNextSectionID(c *echo.Context) error {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return apierror.JSON(c, http.StatusBadRequest, "invalid site id")
	}

	ctx := c.Request().Context()

	_, err = h.queries.GetSiteByID(ctx, siteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.JSON(c, http.StatusNotFound, "site not found")
		}
		return apierror.Internal(c, "failed to get site", err)
	}

	id, err := h.queries.GetNextSectionID(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get next section id", err)
	}

	return c.JSON(http.StatusOK, NextSectionIDResponse{ID: id})
}
