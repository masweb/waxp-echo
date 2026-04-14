package handler

import (
	"github.com/labstack/echo/v5"

	"waxp/echo/internal/apierror"
)

func ErrorJSON(c *echo.Context, code int, msg string, err ...error) error {
	return apierror.JSON(c, code, msg, err...)
}

func InternalError(c *echo.Context, msg string, err error) error {
	return apierror.Internal(c, msg, err)
}
