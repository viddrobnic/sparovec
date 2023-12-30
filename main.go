package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/database"
	"github.com/viddrobnic/sparovec/features/auth"
	"github.com/viddrobnic/sparovec/features/dashboard"
	"github.com/viddrobnic/sparovec/features/tags"
	"github.com/viddrobnic/sparovec/features/transactions"
	"github.com/viddrobnic/sparovec/features/wallets"
	"github.com/viddrobnic/sparovec/observability"
)

//go:embed config.toml
var defaultConfig []byte

//go:embed migrations
var migrationsDir embed.FS

//go:embed assets/*
var assetsDir embed.FS

func main() {
	err := config.WriteDefault(defaultConfig)
	if err != nil {
		log.Fatal(err)
	}

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
	usersRepository := auth.NewRepository(db)
	usersService := auth.New(usersRepository, nil, logger)

	user, err := usersService.CreateUser(context.Background(), username, password)
	if err != nil {
		_, _ = fmt.Println("Failed to create user")
		return
	}

	_, _ = fmt.Printf("User created: %d\n", user.Id)
}

func serve(conf *config.Config, db *sqlx.DB, logger *slog.Logger) {
	usersRepository := auth.NewRepository(db)
	walletsRepository := wallets.NewRepository(db)
	tagsRepository := tags.NewRepository(db)
	transactionRepository := transactions.NewRepository(db)
	dashboardRepository := dashboard.NewRepository(db)

	authRoutes := auth.New(usersRepository, conf, logger)
	walletsRoutes := wallets.New(
		walletsRepository,
		usersRepository,
		logger.With("where", "wallets_routes"),
	)
	dashboardRoutes := dashboard.New(
		dashboardRepository,
		walletsRepository,
		logger.With("where", "dashboard_routes"),
	)
	tagsRoutes := tags.New(
		walletsRepository,
		tagsRepository,
		logger.With("where", "tags_routes"),
	)
	transactionsRoutes := transactions.New(
		transactionRepository,
		tagsRepository,
		walletsRepository,
		logger.With("where", "transactions_routes"),
	)

	router := createRouter(conf, authRoutes)
	staticFs, _ := fs.Sub(assetsDir, "assets")
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))))

	authRoutes.Mount(router)
	walletsRoutes.Mount(router)
	dashboardRoutes.Mount(router)
	tagsRoutes.Mount(router)
	transactionsRoutes.Mount(router)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", conf.API.ListenAddress, conf.API.Port), router)
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
	err = database.Migrate(db, migrationsDir)
	if err != nil {
		logger.Error("Failed to migrate database", "error", err)
		return nil, err
	}

	return db, nil
}

func createRouter(conf *config.Config, authService auth.Service) chi.Router {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.StripSlashes,
		cors.Handler(cors.Options{
			AllowedOrigins:   conf.API.CorsAllowedOrigins,
			AllowCredentials: true,
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders: []string{"*"},
		}),
	)

	if conf.Observability.WriteToConsole {
		router.Use(middleware.Logger)
	}

	router.Use(
		middleware.Timeout(5*time.Second),
		auth.CreateMiddleware(authService),
		middleware.Recoverer,
	)

	return router
}
