package handler

import (
	"github.com/labstack/echo/v5"
)

type APIError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func ErrorJSON(c *echo.Context, code int, msg string) error {
	return c.JSON(code, APIError{Error: msg, Code: code})
}
