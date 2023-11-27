package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type TransactionService interface {
	Create(ctx context.Context, transaction *models.Transaction, user *models.User) error
	Update(ctx context.Context, transaction *models.Transaction, user *models.User) error
	List(ctx context.Context, req *models.TransactionsListRequest, user *models.User) (*models.PaginatedResponse[*models.Transaction], error)
	Delete(ctx context.Context, id, walletId int, user *models.User) error
}

type TransactionTagsService interface {
	List(ctx context.Context, walletId int, user *models.User) ([]*models.Tag, error)
}

type Transactions struct {
	navbarService      NavbarWalletsService
	transactionService TransactionService
	tagsService        TransactionTagsService
	log                *slog.Logger

	// Templates
	transactionsTemplate *template.Template
}

func NewTransactions(
	navbarService NavbarWalletsService,
	transactionService TransactionService,
	tagsService TransactionTagsService,
	templates fs.FS,
	log *slog.Logger,
) *Transactions {
	transactionsTemplate := template.Must(template.ParseFS(
		templates,
		"templates/index.html",
		"templates/layout.html",
		"templates/transactions/transactions.html",
		"templates/transactions/components/*",
	))

	return &Transactions{
		navbarService:      navbarService,
		transactionService: transactionService,
		tagsService:        tagsService,
		log:                log,

		transactionsTemplate: transactionsTemplate,
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

type listTransactionsForm struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func listTransactionFormFromRequest(r *http.Request) *listTransactionsForm {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	return &listTransactionsForm{
		Page:     page,
		PageSize: pageSize,
	}
}

func (t *Transactions) transactions(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	walletId := getWalletId(r)

	form := listTransactionFormFromRequest(r)

	req := &models.TransactionsListRequest{
		WalletId: walletId,
		Page:     models.NewPage(form.Page, form.PageSize),
	}

	// Get transactions
	paginatedTransactions, err := t.transactionService.List(r.Context(), req, user)
	if err != nil {
		t.log.Error("Failed to list transactions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get tags
	tags, err := t.tagsService.List(r.Context(), walletId, user)
	if err != nil {
		t.log.Error("Failed to list tags", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get navbar context
	navbarCtx, err := createNavbarContext(r, t.navbarService)
	if err != nil {
		t.log.ErrorContext(r.Context(), "Failed to create navbar context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate number of pages
	pages := int(math.Ceil(float64(paginatedTransactions.Count) / float64(req.Page.PageSize)))

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

	ctx := &models.TransactionsContext{
		Navbar:          navbarCtx,
		Transactions:    models.RenderTransactions(paginatedTransactions.Data),
		Tags:            tags,
		CurrentPage:     req.Page.Page,
		TotalPages:      pages,
		UrlParams:       r.URL.RawQuery,
		PreviousPageUrl: prevUrl,
		NextPageUrl:     nextUrl,
	}

	err = t.transactionsTemplate.Execute(w, ctx)
	if err != nil {
		t.log.Error("Failed to render template", "error", err)
	}
}

type transactionFormSubmitType string

const (
	transactionFormSubmitTypeCreate transactionFormSubmitType = "create"
	transactionFormSubmitTypeEdit   transactionFormSubmitType = "update"
)

type transactionType string

const (
	transactionTypeOutcome transactionType = "outcome"
	transactionTypeIncome  transactionType = "income"
)

type saveTransactionForm struct {
	Id         int                       `form:"id"`
	SubmitType transactionFormSubmitType `form:"submit_type"`
	Name       string                    `form:"name"`
	Type       transactionType           `form:"type"`
	Value      string                    `form:"value"`
	TagId      string                    `form:"tag"`
	Date       string                    `form:"date"`
}

func (f *saveTransactionForm) parse(walletId int) (*models.Transaction, error) {
	// Parse value
	valueStr := strings.ReplaceAll(f.Value, ",", ".")
	valueF, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, &models.ErrInvalidForm{Message: "Value is not a number"}
	}

	value := int(math.Round(valueF * 100))
	if value < 0 {
		return nil, &models.ErrInvalidForm{Message: "Value must be positive"}
	}

	if f.Type == transactionTypeOutcome {
		value *= -1
	}

	// Parse date
	date, err := time.Parse("2006-01-02", f.Date)
	if err != nil {
		return nil, &models.ErrInvalidForm{Message: "Invalid date"}
	}

	// Parse tag
	var tag *models.Tag
	if f.TagId != "" {
		tagId, err := strconv.Atoi(f.TagId)
		if err != nil {
			return nil, &models.ErrInvalidForm{Message: "Invalid tag"}
		}

		tag = &models.Tag{Id: tagId}
	}

	return &models.Transaction{
		Id:        f.Id,
		WalletId:  walletId,
		Name:      f.Name,
		Value:     value,
		CreatedAt: date,
		Tag:       tag,
	}, nil
}

func saveTransactionFormFromRequest(r *http.Request) *saveTransactionForm {
	id, _ := strconv.Atoi(r.FormValue("id"))
	submitType := transactionFormSubmitType(r.FormValue("submit_type"))
	transactionType := transactionType(r.FormValue("type"))

	return &saveTransactionForm{
		Id:         id,
		SubmitType: submitType,
		Name:       r.FormValue("name"),
		Type:       transactionType,
		Value:      r.FormValue("value"),
		TagId:      r.FormValue("tag"),
		Date:       r.FormValue("date"),
	}
}

func (t *Transactions) saveTransaction(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	walletId := getWalletId(r)

	form := saveTransactionFormFromRequest(r)
	transaction, err := form.parse(walletId)
	if err != nil {
		t.handleError(w, err)
		return
	}

	switch form.SubmitType {
	case transactionFormSubmitTypeCreate:
		err = t.transactionService.Create(r.Context(), transaction, user)
	case transactionFormSubmitTypeEdit:
		err = t.transactionService.Update(r.Context(), transaction, user)
	default:
		t.log.Error("Invalid submit type", "submit_type", form.SubmitType)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err != nil {
		t.handleError(w, err)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventSaveSuccess)
	t.transactions(w, r)
}

func (t *Transactions) deleteTransaction(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	walletId := getWalletId(r)

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		t.log.Error("Failed to parse transaction id", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.transactionService.Delete(r.Context(), id, walletId, user)
	if err != nil {
		t.log.Error("Failed to delete transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventDeleteSuccess)
	t.transactions(w, r)
}

func (t *Transactions) handleError(w http.ResponseWriter, err error) {
	var invalidForm *models.ErrInvalidForm
	if errors.As(err, &invalidForm) {
		saveError := HtmxEventSaveError{ErrorMessage: invalidForm.Message}
		saveErrorJson, err := json.Marshal(saveError)
		if err != nil {
			t.log.Error("Failed to marshal save error", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set(HtmxHeaderReswap, HtmxSwapNone)
		w.Header().Set(HtmxHeaderTriggerAfterSettle, string(saveErrorJson))

		w.WriteHeader(http.StatusOK)
	} else {
		t.log.Error("Failed to save transaction", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
