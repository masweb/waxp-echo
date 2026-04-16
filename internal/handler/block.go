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

type BlockHandler struct {
	queries *db.Queries
}

func NewBlockHandler(queries *db.Queries) *BlockHandler {
	return &BlockHandler{queries: queries}
}

type NextBlockIDResponse struct {
	ID int64 `json:"id"`
}

func (h *BlockHandler) GetNextBlockID(c *echo.Context) error {
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

	id, err := h.queries.GetNextBlockID(ctx, siteID)
	if err != nil {
		return apierror.Internal(c, "failed to get next block id", err)
	}

	return c.JSON(http.StatusOK, NextBlockIDResponse{ID: id})
}
