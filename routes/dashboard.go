package routes

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features/auth"
)

type Dashboard struct {
	navbarService NavbarWalletsService
	log           *slog.Logger

	// Templates
	dashboardTemplate *template.Template
}

func NewDashboard(navbarService NavbarWalletsService, templates fs.FS, log *slog.Logger) *Dashboard {
	dashboardTemplate := template.Must(template.ParseFS(
		templates,
		"templates/index.html",
		"templates/layout.html",
		"templates/dashboard/dashboard.html",
	))

	return &Dashboard{
		navbarService: navbarService,
		log:           log,

		dashboardTemplate: dashboardTemplate,
	}
}

func (d *Dashboard) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", d.dashboard)

	router.Mount("/wallets/{walletId}", group)
}

func (d *Dashboard) dashboard(w http.ResponseWriter, r *http.Request) {
	navbarCtx, err := createNavbarContext(r, d.navbarService)
	if err != nil {
		d.log.Error("Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = renderTemplate(w, d.dashboardTemplate, map[string]any{"Navbar": navbarCtx})
	if err != nil {
		d.log.Error("Failed to render template", "error", err)
	}
}
