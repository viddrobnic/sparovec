package routes

import (
	"context"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type SettingsService interface {
	WalletName(ctx context.Context, walletId int, user *models.User) (string, error)
	Members(ctx context.Context, walletId int, user *models.User) ([]*models.Member, error)
}

type Settings struct {
	navbarService   NavbarWalletsService
	settingsService SettingsService
	log             *slog.Logger

	// Templates
	settingsTemplate *template.Template
}

func NewSettings(
	navbarService NavbarWalletsService,
	settingsService SettingsService,
	templates fs.FS,
	log *slog.Logger,
) *Settings {
	settingsTemplate := template.Must(template.ParseFS(
		templates,
		"templates/index.html",
		"templates/layout.html",
		"templates/settings/settings.html",
		"templates/settings/components/*",
	))

	return &Settings{
		navbarService:   navbarService,
		settingsService: settingsService,
		log:             log,

		settingsTemplate: settingsTemplate,
	}
}

func (s *Settings) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", s.settings)

	router.Mount("/wallets/{walletId}/settings", group)
}

func (s *Settings) settings(w http.ResponseWriter, r *http.Request) {
	navbarCtx, err := createNavbarContext(r, s.navbarService)
	if err != nil {
		s.log.Error("Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx := &models.SettingsContext{
		Navbar:     navbarCtx,
		WalletName: "My Wallet",
		Members: []models.Member{
			{Id: 1, Username: "John", IsSelf: true},
			{Id: 2, Username: "Jane"},
		},
	}

	err = s.settingsTemplate.Execute(w, ctx)
	if err != nil {
		s.log.Error("Failed to execute template", "error", err)
	}
}
