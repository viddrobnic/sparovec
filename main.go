package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/database"
	"github.com/viddrobnic/sparovec/observability"
	"github.com/viddrobnic/sparovec/repository"
	"github.com/viddrobnic/sparovec/service"
)

func main() {
	conf, err := config.LoadDefault()
	if err != nil {
		log.Fatal(err)
	}

	logger := observability.NewLogger(conf)

	db, err := setupDatabase(conf, logger)
	if err != nil {
		return
	}

	if len(os.Args) < 2 {
		_, _ = fmt.Println("Usage: sparovec <command>")
		_, _ = fmt.Println("\tserve\t\t\t\t\tStarts the server")
		_, _ = fmt.Println("\tcreate-user [username] [password]\tCreates a new user with given credentials")
		return
	}

	switch os.Args[1] {
	case "serve":
		serve(conf, db, logger)
	case "create-user":
		if len(os.Args) != 4 {
			_, _ = fmt.Println("Usage: sparovec create-user [username] [password]")
			return
		}

		createUser(db, logger, os.Args[2], os.Args[3])
	default:
		_, _ = fmt.Println("Unknown command")
	}
}

func createUser(db *sqlx.DB, log *slog.Logger, username, password string) {
	usersRepository := repository.NewUsers(db)
	usersService := service.NewUser(usersRepository, log)

	user, err := usersService.Create(context.Background(), username, password)
	if err != nil {
		_, _ = fmt.Println("Failed to create user")
		return
	}

	_, _ = fmt.Printf("User created: %d\n", user.Id)
}

func serve(conf *config.Config, db *sqlx.DB, logger *slog.Logger) {
	router := createRouter(conf)

	router.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "asdf")
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

	err := router.Start(fmt.Sprintf("%s:%d", conf.API.ListenAddress, conf.API.Port))
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

func createRouter(conf *config.Config) *echo.Echo {
	router := echo.New()
	router.Use(
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: 5 * time.Second,
		}),
		middleware.RequestID(),
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

	return router
}
