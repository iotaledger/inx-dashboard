package common

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	// ErrInvalidParameter defines the invalid parameter error.
	ErrInvalidParameter = echo.NewHTTPError(http.StatusBadRequest, "invalid parameter")
)
