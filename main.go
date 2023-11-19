package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/database"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/observability"
	"github.com/viddrobnic/sparovec/repository"
	"github.com/viddrobnic/sparovec/routes"
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

func createUser(db *sqlx.DB, logger *slog.Logger, username, password string) {
	usersRepository := repository.NewUsers(db)
	usersService := service.NewUser(usersRepository, logger)

	user, err := usersService.Create(context.Background(), username, password)
	if err != nil {
		_, _ = fmt.Println("Failed to create user")
		return
	}

	_, _ = fmt.Printf("User created: %d\n", user.Id)
}

func serve(conf *config.Config, db *sqlx.DB, logger *slog.Logger) {
	usersRepository := repository.NewUsers(db)
	walletsRepository := repository.NewWallets(db)

	authService := service.NewAuth(usersRepository, conf, logger)

	authRoutes := routes.NewAuth(authService)
	walletsRoutes := routes.NewWallets(walletsRepository, logger)
	dashboardRoutes := routes.NewDashboard(walletsRepository, logger)

	router := createRouter(conf, authService)
	router.Static("/static", "assets")

	authRoutes.Mount(router.Group("/auth"))
	walletsRoutes.Mount(router.Group(""))
	dashboardRoutes.Mount(router.Group("/wallets/:walletId"))

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

func createRouter(conf *config.Config, authService auth.Service) *echo.Echo {
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusTemporaryRedirect,
	}))

	router.Use(
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: 5 * time.Second,
		}),
		middleware.RequestID(),
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
		auth.CreateMiddleware(authService),
		middleware.Recover(),
	)

	return router
}
