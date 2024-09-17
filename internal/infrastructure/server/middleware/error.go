package middleware

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"tenderSystem/internal/domain"
)

func NewErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			fmt.Println(err)

			if err == nil {
				return nil
			}

			if errors.Is(err, domain.ErrInvalidArgument) {
				return echo.NewHTTPError(400, err.Error())
			}

			if errors.Is(err, domain.ErrUnauthorized) {
				return echo.NewHTTPError(401, err.Error())
			}

			if errors.Is(err, domain.ErrForbidden) {
				return echo.NewHTTPError(403, err.Error())
			}

			if errors.Is(err, domain.ErrNotFound) {
				return echo.NewHTTPError(404, err.Error())
			}

			if errors.Is(err, domain.ErrAlreadyExists) {
				return echo.NewHTTPError(409, err.Error())
			}

			return echo.NewHTTPError(500, err.Error())
		}
	}
}
