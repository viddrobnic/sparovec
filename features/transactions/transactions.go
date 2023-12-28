package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/features"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
	"github.com/viddrobnic/sparovec/routes"
)

type Repository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	Update(ctx context.Context, transaction *models.Transaction) error
	List(ctx context.Context, req *models.TransactionsListRequest) ([]*models.Transaction, int, error)
	Delete(ctx context.Context, id int) error
}

type TagsRepository interface {
	List(ctx context.Context, walletId int) ([]*models.Tag, error)
	Get(ctx context.Context, tagId int) (*models.Tag, error)
	GetIds(ctx context.Context, tagIds []int) ([]*models.Tag, error)
}

type WalletRepository interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
}

type Transactions struct {
	repository       Repository
	tagsRepository   TagsRepository
	walletRepository WalletRepository

	log *slog.Logger
}

func New(
	repository Repository,
	tagsRepository TagsRepository,
	walletRepository WalletRepository,
	log *slog.Logger,
) *Transactions {
	return &Transactions{
		repository:       repository,
		tagsRepository:   tagsRepository,
		walletRepository: walletRepository,

		log: log,
	}
}

func (t *Transactions) Mount(router chi.Router) {
	group := chi.NewRouter()
	group.Use(auth.RequiredMiddleware)

	group.Get("/", t.transactions)
	group.Post("/", t.saveTransaction)
	group.Post("/delete", t.deleteTransaction)

	router.Mount("/wallets/{walletId}/transactions", group)
}

func (t *Transactions) hasPermission(ctx context.Context, w http.ResponseWriter, walletId, userId int) bool {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, userId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func (t *Transactions) transactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if !t.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	form := listTransactionFormFromRequest(r)
	req := &models.TransactionsListRequest{
		WalletId: walletId,
		Page:     models.NewPage(form.Page, form.PageSize),
	}

	transactions, count, err := t.repository.List(ctx, req)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to list transactions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = t.expandTags(ctx, transactions)
	if err != nil {
		t.log.WarnContext(ctx, "Failed to expand tags", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tags, err := t.tagsRepository.List(ctx, walletId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to list tags", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	wallets, err := t.walletRepository.ForUser(ctx, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate number of pages
	pages := int(math.Ceil(float64(count) / float64(req.Page.PageSize)))

	// Get previous and next page
	var prevUrl, nextUrl string
	query := r.URL.Query()
	page := req.Page.Page
	if page > 1 {
		query.Set("page", strconv.Itoa(page-1))
		prevUrl = fmt.Sprintf("/wallets/%d/transactions?%s", walletId, query.Encode())
	}
	if page < pages {
		query.Set("page", strconv.Itoa(page+1))
		nextUrl = fmt.Sprintf("/wallets/%d/transactions?%s", walletId, query.Encode())
	}

	navbar := models.Navbar{
		SelectedWalletId: walletId,
		Wallets:          wallets,
		Username:         user.Username,
		Title:            "Šparovec | Transactions",
	}

	view := transactionsView(transactionsViewData{
		navbar:          navbar,
		transactions:    models.RenderTransactions(transactions),
		tags:            tags,
		currentPage:     strconv.Itoa(page),
		totalPages:      strconv.Itoa(pages),
		previousPageUrl: templ.SafeURL(prevUrl),
		nextPageUrl:     templ.SafeURL(nextUrl),
		urlParams:       r.URL.RawQuery,
	})
	err = view.Render(ctx, w)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to render view", "error", err)
	}
}

func (t *Transactions) saveTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if !t.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	form := saveTransactionFormFromRequest(r)
	transaction, err := form.parse(walletId)
	if err != nil {
		t.handleError(w, err)
		return
	}

	if err := t.validateTag(ctx, transaction.Tag, transaction.WalletId); err != nil {
		t.handleError(w, err)
		return
	}

	switch form.SubmitType {
	case transactionFormSubmitTypeCreate:
		err = t.repository.Create(r.Context(), transaction)
	case transactionFormSubmitTypeEdit:
		err = t.repository.Update(r.Context(), transaction)
	default:
		t.log.Error("Invalid submit type", "submit_type", form.SubmitType)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err != nil {
		t.handleError(w, err)
		return
	}

	w.Header().Set(routes.HtmxHeaderTriggerAfterSettle, routes.HtmxEventSaveSuccess)
	t.transactions(w, r)
}

func (t *Transactions) deleteTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(r)
	walletId := features.GetWalletId(r)

	if !t.hasPermission(ctx, w, walletId, user.Id) {
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		t.log.Error("Failed to parse transaction id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.repository.Delete(ctx, id)
	if err != nil {
		t.log.Error("Failed to delete transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(routes.HtmxHeaderTriggerAfterSettle, routes.HtmxEventDeleteSuccess)
	t.transactions(w, r)
}

func (t *Transactions) handleError(w http.ResponseWriter, err error) {
	var invalidForm *models.ErrInvalidForm
	if errors.As(err, &invalidForm) {
		saveError := routes.HtmxEventSaveError{ErrorMessage: invalidForm.Message}
		saveErrorJson, err := json.Marshal(saveError)
		if err != nil {
			t.log.Error("Failed to marshal save error", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set(routes.HtmxHeaderReswap, routes.HtmxSwapNone)
		w.Header().Set(routes.HtmxHeaderTriggerAfterSettle, string(saveErrorJson))

		w.WriteHeader(http.StatusOK)
	} else {
		t.log.Error("Failed to save transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (t *Transactions) validateTag(ctx context.Context, tag *models.Tag, walletId int) error {
	if tag == nil {
		return nil
	}

	tag, err := t.tagsRepository.Get(ctx, tag.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tag", "error", err)
		return models.ErrInternalServer
	}

	if tag == nil {
		return &models.ErrInvalidForm{Message: "Invalid tag"}
	}

	if tag.WalletId != walletId {
		return &models.ErrInvalidForm{Message: "Invalid tag"}
	}

	return nil
}

func (t *Transactions) expandTags(ctx context.Context, transactions []*models.Transaction) error {
	ids := make([]int, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.Tag != nil {
			ids = append(ids, transaction.Tag.Id)
		}
	}

	tags, err := t.tagsRepository.GetIds(ctx, ids)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tags", "error", err)
		return models.ErrInternalServer
	}

	tagsMap := make(map[int]*models.Tag)
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	for _, transaction := range transactions {
		if transaction.Tag != nil {
			transaction.Tag = tagsMap[transaction.Tag.Id]
		}
	}

	return nil
}
