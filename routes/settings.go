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
	ChangeWalletName(ctx context.Context, walletId int, name string, user *models.User) error
	AddMember(ctx context.Context, walletId int, username string, user *models.User) error
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
	group.Post("/name", s.saveName)
	group.Post("/add-member", s.addMember)

	router.Mount("/wallets/{walletId}/settings", group)
}

func (s *Settings) settings(w http.ResponseWriter, r *http.Request) {
	walletid := getWalletId(r)
	user := auth.GetUser(r)

	navbarCtx, err := createNavbarContext(r, s.navbarService)
	if err != nil {
		s.log.Error("Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	name, err := s.settingsService.WalletName(r.Context(), walletid, user)
	if err != nil {
		s.log.Error("Failed to get wallet name", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	members, err := s.settingsService.Members(r.Context(), walletid, user)
	if err != nil {
		s.log.Error("Failed to get wallet members", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx := &models.SettingsContext{
		Navbar:     navbarCtx,
		WalletName: name,
		Members:    members,
	}

	err = s.settingsTemplate.Execute(w, ctx)
	if err != nil {
		s.log.Error("Failed to execute template", "error", err)
	}
}

func (s *Settings) saveName(w http.ResponseWriter, r *http.Request) {
	walletId := getWalletId(r)
	user := auth.GetUser(r)
	name := r.FormValue("name")

	err := s.settingsService.ChangeWalletName(r.Context(), walletId, name, user)
	if err != nil {
		s.log.Error("Failed to change wallet name", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.settings(w, r)
}

func (s *Settings) addMember(w http.ResponseWriter, r *http.Request) {
	walletId := getWalletId(r)
	user := auth.GetUser(r)
	username := r.FormValue("username")

	err := s.settingsService.AddMember(r.Context(), walletId, username, user)
	if err != nil {
		s.log.Error("Failed to add user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.settings(w, r)
}
