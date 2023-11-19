package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/models"
)

type Service interface {
	ValidateSession(session *models.Session) error
}

func CreateMiddleware(service Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(models.SessionCookieName)
			if err != nil {
				return next(c)
			}

			session, err := models.SessionFromCookie(cookie.Value)
			if err != nil {
				return next(c)
			}

			if err := service.ValidateSession(session); err != nil {
				return next(c)
			}

			c.Set(models.UserContextKey, session.User)
			return next(c)
		}
	}
}

func RequiredMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get(models.UserContextKey)
		if user == nil {
			return c.Redirect(http.StatusSeeOther, "/auth/sign-in")
		}

		if _, ok := user.(*models.User); !ok {
			return c.Redirect(http.StatusSeeOther, "/auth/sign-in")
		}

		return next(c)
	}
}
