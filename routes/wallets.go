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

type WalletsService interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
	Create(ctx context.Context, userId int, name string) (*models.Wallet, error)
}

type Wallets struct {
	service WalletsService
	log     *slog.Logger

	// Templates
	walletsTemplate    *template.Template
	walletCardTemplate *template.Template
}

func NewWallets(service WalletsService, templates fs.FS, log *slog.Logger) *Wallets {
	walletsTemplate := template.Must(template.ParseFS(
		templates,
		"templates/index.html",
		"templates/layout.html",
		"templates/wallets/wallets.html",
		"templates/wallets/components/*",
	))

	return &Wallets{
		service: service,
		log:     log,

		walletsTemplate: walletsTemplate,
	}
}

func (wlts *Wallets) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", wlts.wallets)
	group.Post("/", wlts.createWallet)

	router.Mount("/", group)
}

func (wlts *Wallets) wallets(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	navbarCtx, err := createNavbarContext(r, wlts.service)
	if err != nil {
		wlts.log.Error("Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wallets, err := wlts.service.ForUser(r.Context(), user.Id)
	if err != nil {
		wlts.log.Error("Failed to get wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx := &models.WalletsContext{
		Navbar:  navbarCtx,
		Wallets: wallets,
	}

	err = renderTemplate(w, wlts.walletsTemplate, ctx)
	if err != nil {
		wlts.log.Error("Failed to render template", "error", err)
	}
}

func (wlts Wallets) createWallet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	name := r.FormValue("name")

	wallet, err := wlts.service.Create(r.Context(), user.Id, name)
	if err != nil {
		wlts.log.Error("Failed to create wallet", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventCreateSuccess)
	err = renderTemplateNamed(w, wlts.walletCardTemplate, "wallet-card", wallet)
	if err != nil {
		wlts.log.Error("Failed to render template", "error", err)
	}
}
