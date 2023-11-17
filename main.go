package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/database"
	"github.com/viddrobnic/sparovec/observability"
)

func main() {
	conf, err := config.LoadDefault()
	if err != nil {
		log.Fatal(err)
	}

	logger := observability.NewLogger(conf)

	_, err = setupDatabase(conf, logger)
	if err != nil {
		return
	}

	// Create router
	router := echo.New()
	router.Use(middleware.RequestID(),
		middleware.RemoveTrailingSlash(),
		middleware.Logger(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     conf.API.CorsAllowedOrigins,
			AllowCredentials: true,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowHeaders: []string{"*"},
		}),
		middleware.Secure(),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if strings.HasPrefix(c.Request().URL.Path, "/auth") {
					return next(c)
				}

				cookie, err := c.Cookie("session")
				if err != nil {
					return c.String(http.StatusUnauthorized, "Unauthorized")
				}

				if cookie.Value != "asdf" {
					return c.String(http.StatusUnauthorized, "Unauthorized")
				}

				return next(c)
			}
		},
		middleware.Recover(),
	)

	router.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	router.GET("/auth/login", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    "asdf",
			Path:     "/",
			MaxAge:   10 * 24 * 60 * 60,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(cookie)

		return c.Redirect(http.StatusSeeOther, "/")
	})

	router.GET("/auth/logout", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(cookie)

		return c.Redirect(http.StatusSeeOther, "/")
	})

	err = router.Start(fmt.Sprintf("%s:%d", conf.API.ListenAddress, conf.API.Port))
	if err != nil {
		logger.Error("Failed to start server", "error", err)
		return
	}
}

func setupDatabase(conf *config.Config, logger *slog.Logger) (*sqlx.DB, error) {
	logger.Info("Connecting to database")
	db, err := database.New(conf)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	logger.Info("Migrating database")
	err = database.Migrate(db)
	if err != nil {
		logger.Error("Failed to migrate database", "error", err)
		return nil, err
	}

	return db, nil
}
