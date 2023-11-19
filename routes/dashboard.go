package routes

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Dashboard struct {
	navbarService NavbarWalletsService
	log           *slog.Logger

	// Templates
	dashboardTemplate *template.Template
}

func NewDashboard(navbarService NavbarWalletsService, log *slog.Logger) *Dashboard {
	dashboardTemplate := template.Must(template.ParseFiles(
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

func (d *Dashboard) Mount(group *echo.Group) {
	group.GET("", d.dashboard)

	group.RouteNotFound("/*", func(c echo.Context) error {
		// TODO: Better not found
		return c.NoContent(http.StatusNotFound)
	})
}

func (d *Dashboard) dashboard(c echo.Context) error {
	navbarCtx, err := createNavbarContext(c, d.navbarService)
	if err != nil {
		d.log.Error("Failed to create navbar context", "error", err)
		return err
	}

	return renderTemplate(c, d.dashboardTemplate, map[string]any{"Navbar": navbarCtx})
}
