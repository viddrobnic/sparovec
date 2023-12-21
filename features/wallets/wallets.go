package wallets

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type RepositoryInterface interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
}

type Wallets struct {
	repository RepositoryInterface
	log        *slog.Logger
}

func New(repository RepositoryInterface, log *slog.Logger) *Wallets {
	return &Wallets{
		repository: repository,
		log:        log,
	}
}

func (wlts *Wallets) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", wlts.wallets)

	// group.Get("/", wlts.wallets)
	// group.Post("/", wlts.createWallet)

	router.Mount("/", group)
}

func (wlts *Wallets) wallets(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	wallets, err := wlts.repository.ForUser(r.Context(), user.Id)
	if err != nil {
		wlts.log.Error("Failed to get wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	navbar := models.Navbar{
		SelectedWalletId: features.GetWalletId(r),
		Wallets:          wallets,
		Username:         user.Username,
		Title:            "Šparovec",
	}

	view := WalletsView(wallets, navbar)
	err = view.Render(r.Context(), w)
	if err != nil {
		wlts.log.Error("Failed to render view", "error", err)
	}
}

// func (wlts *Wallets) wallets(w http.ResponseWriter, r *http.Request) {
// 	user := auth.GetUser(r)
//
// 	navbarCtx, err := createNavbarContext(r, wlts.service)
// 	if err != nil {
// 		wlts.log.Error("Failed to create navbar context", "error", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	wallets, err := wlts.service.ForUser(r.Context(), user.Id)
// 	if err != nil {
// 		wlts.log.Error("Failed to get wallets", "error", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	ctx := &models.WalletsContext{
// 		Navbar:  navbarCtx,
// 		Wallets: wallets,
// 	}
//
// 	err = renderTemplate(w, wlts.walletsTemplate, ctx)
// 	if err != nil {
// 		wlts.log.Error("Failed to render template", "error", err)
// 	}
// }
//
// func (wlts Wallets) createWallet(w http.ResponseWriter, r *http.Request) {
// 	user := auth.GetUser(r)
// 	name := r.FormValue("name")
//
// 	wallet, err := wlts.service.Create(r.Context(), user.Id, name)
// 	if err != nil {
// 		wlts.log.Error("Failed to create wallet", "error", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventCreateSuccess)
// 	err = renderTemplateNamed(w, wlts.walletCardTemplate, "wallet-card", wallet)
// 	if err != nil {
// 		wlts.log.Error("Failed to render template", "error", err)
// 	}
// }
