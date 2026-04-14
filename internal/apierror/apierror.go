package apierror

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Response struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func JSON(c *echo.Context, code int, msg string, err ...error) error {
	if code >= 500 {
		attrs := []slog.Attr{
			slog.Int("status", code),
			slog.String("path", c.Request().URL.Path),
			slog.String("method", c.Request().Method),
			slog.String("message", msg),
		}
		if len(err) > 0 && err[0] != nil {
			attrs = append(attrs, slog.String("error", err[0].Error()))
		}
		slog.LogAttrs(c.Request().Context(), slog.LevelError, "internal error", attrs...)
	}

	return c.JSON(code, Response{Error: msg, Code: code})
}

func Internal(c *echo.Context, msg string, err error) error {
	return JSON(c, http.StatusInternalServerError, msg, err)
}
