package dashboard

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features"
	"github.com/viddrobnic/sparovec/features/auth"
	"github.com/viddrobnic/sparovec/models"
)

type Repository interface {
	GetTransactions(ctx context.Context, walletId, year int, mont time.Month) ([]*models.Transaction, error)
}

type WalletRepository interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
}

type Dashboard struct {
	repository       Repository
	walletRepository WalletRepository

	log *slog.Logger
}

func New(
	repository Repository,
	walletRepository WalletRepository,
	log *slog.Logger,
) *Dashboard {
	return &Dashboard{
		repository:       repository,
		walletRepository: walletRepository,

		log: log,
	}
}

func (d *Dashboard) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", d.dashboard)

	router.Mount("/wallets/{walletId}", group)
}

func (d *Dashboard) hasPermission(ctx context.Context, w http.ResponseWriter, walletId, userId int) bool {
	hasPermission, err := d.walletRepository.HasPermission(ctx, walletId, userId)
	if err != nil {
		d.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func (d *Dashboard) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletId := features.GetWalletId(r)
	user := auth.GetUser(r)

	if !d.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	form := dashboardFormFromRequest(r)

	wallets, err := d.walletRepository.ForUser(ctx, user.Id)
	if err != nil {
		d.log.ErrorContext(ctx, "Failed to get user wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = d.repository.GetTransactions(ctx, walletId, form.year, form.month)
	if err != nil {
		d.log.ErrorContext(ctx, "Failed to get transactions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := dashboardViewData{
		navbar: models.Navbar{
			SelectedWalletId: walletId,
			Wallets:          wallets,
			Username:         user.Username,
			Title:            "Šparovec | Dashboard",
		},
		month:   form.month,
		year:    form.year,
		maxYear: time.Now().Year(),
		dType:   form.dType,
	}
	view := dashboardView(data)
	err = view.Render(ctx, w)
	if err != nil {
		d.log.ErrorContext(ctx, "Error rendering dashboard view", "error", err)
	}
}

type dashboardForm struct {
	year  int
	month time.Month
	dType models.DashboardType
}

func dashboardFormFromRequest(r *http.Request) dashboardForm {
	form := dashboardForm{
		year:  time.Now().Year(),
		month: time.Now().Month(),
		dType: models.DashboardTypeExpense,
	}

	yearStr := r.FormValue("year")
	if yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			form.year = year
		}
	}

	monthStr := r.FormValue("month")
	if monthStr != "" {
		month, err := strconv.Atoi(monthStr)
		if err == nil {
			form.month = time.Month(month)
		}
	}

	dType := r.FormValue("type")
	if dType != "" {
		form.dType = models.DashboardType(dType)
	}

	return form
}
