package routes

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/middleware/auth"
)

type Settings struct {
	navbarService NavbarWalletsService
	log           *slog.Logger

	// Templates
	settingsTemplate *template.Template
}

func NewSettings(navbarService NavbarWalletsService, templates fs.FS, log *slog.Logger) *Settings {
	settingsTemplate := template.Must(template.ParseFS(
		templates,
		"templates/index.html",
		"templates/layout.html",
		"templates/settings/settings.html",
		"templates/settings/components/*",
	))

	return &Settings{
		navbarService: navbarService,
		log:           log,

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

	err = s.settingsTemplate.Execute(w, map[string]any{"Navbar": navbarCtx})
	if err != nil {
		s.log.Error("Failed to execute template", "error", err)
	}
}
