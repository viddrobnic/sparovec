package dashboard

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
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

type TransactionsService interface {
	ExpandTags(ctx context.Context, transactions []*models.Transaction) error
}

type Dashboard struct {
	repository          Repository
	walletRepository    WalletRepository
	transactionsService TransactionsService

	log *slog.Logger
}

func New(
	repository Repository,
	walletRepository WalletRepository,
	transactionsService TransactionsService,
	log *slog.Logger,
) *Dashboard {
	return &Dashboard{
		repository:          repository,
		walletRepository:    walletRepository,
		transactionsService: transactionsService,

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

	transactions, err := d.repository.GetTransactions(ctx, walletId, form.year, form.month)
	if err != nil {
		d.log.ErrorContext(ctx, "Failed to get transactions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = d.transactionsService.ExpandTags(ctx, transactions)
	if err != nil {
		d.log.ErrorContext(ctx, "Failed to expand tags", "error", err)
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
		data:    createDashboardData(transactions),
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
}

func dashboardFormFromRequest(r *http.Request) dashboardForm {
	form := dashboardForm{
		year:  time.Now().Year(),
		month: time.Now().Month(),
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

	return form
}

func createDashboardData(transactions []*models.Transaction) models.DashboardData {
	data := models.DashboardData{
		NrTransactions: len(transactions),
		TagBalance:     []models.TagBalance{},
	}

	tagBalance := make(map[int]int)
	tags := make(map[int]*models.Tag)

	for _, tr := range transactions {
		if tr.Value > 0 {
			data.Income += tr.Value
		} else {
			data.Outcome += tr.Value

			if tr.Tag != nil {
				tagBalance[tr.Tag.Id] += tr.Value
				tags[tr.Tag.Id] = tr.Tag
			} else {
				tagBalance[0] += tr.Value
				tags[0] = &models.Tag{
					Name:     "Other",
					WalletId: tr.WalletId,
				}
			}
		}

		data.Balance += tr.Value
	}

	for tagId, balance := range tagBalance {
		data.TagBalance = append(data.TagBalance, models.TagBalance{
			Tag:     tags[tagId],
			Balance: balance,
		})
	}

	sort.Slice(data.TagBalance, func(i, j int) bool {
		b1 := data.TagBalance[i]
		b2 := data.TagBalance[j]

		if b1.Balance == b2.Balance {
			return b1.Tag.Name < b2.Tag.Name
		}

		return b1.Balance < b2.Balance
	})

	return data
}
