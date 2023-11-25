package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type TransactionService interface {
	SaveTransaction(ctx context.Context, transaction *models.SaveTransactionForm, walletId int, user *models.User) (*models.Transaction, error)
	List(ctx context.Context, req *models.TransactionsListRequest, user *models.User) (*models.PaginatedResponse[*models.Transaction], error)
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
	log *slog.Logger,
) *Transactions {
	transactionsTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/layout.html",
		"templates/transactions/transactions.html",
	))
	template.Must(transactionsTemplate.ParseGlob("templates/transactions/components/*"))

	return &Transactions{
		navbarService:      navbarService,
		transactionService: transactionService,
		tagsService:        tagsService,
		log:                log,

		transactionsTemplate: transactionsTemplate,
	}
}

func (t *Transactions) Mount(group *echo.Group) {
	group.Use(auth.RequiredMiddleware)

	group.GET("", t.transactions)
	group.POST("", t.saveTransaction)

	group.RouteNotFound("/*", func(c echo.Context) error {
		// TODO: Better not found
		return c.NoContent(http.StatusNotFound)
	})
}

type listTransactionsForm struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func (t *Transactions) transactions(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	walletId := getWalletId(c)

	form := &listTransactionsForm{}
	if err := c.Bind(form); err != nil {
		return err
	}

	return t.renderTransactions(c, form, walletId, user)
}

func (t *Transactions) renderTransactions(c echo.Context, form *listTransactionsForm, walletId int, user *models.User) error {
	req := &models.TransactionsListRequest{
		WalletId: walletId,
		Page:     models.NewPage(form.Page, form.PageSize),
	}

	paginatedTransactions, err := t.transactionService.List(c.Request().Context(), req, user)
	if err != nil {
		return err
	}

	pages := int(math.Ceil(float64(paginatedTransactions.Count) / float64(req.Page.PageSize)))

	tags, err := t.tagsService.List(c.Request().Context(), walletId, user)
	if err != nil {
		return err
	}

	navbarCtx, err := createNavbarContext(c, t.navbarService)
	if err != nil {
		t.log.ErrorContext(c.Request().Context(), "Failed to create navbar context", "error", err)
		return err
	}

	ctx := &models.TransactionsContext{
		Navbar:       navbarCtx,
		Transactions: models.RenderTransactions(paginatedTransactions.Data),
		Tags:         tags,
		CurrentPage:  req.Page.Page,
		PageSize:     req.Page.PageSize,
		TotalPages:   pages,
	}

	fmt.Println(ctx)

	return t.transactionsTemplate.Execute(c.Response().Writer, ctx)
}

type saveTransactionForm struct {
	listTransactionsForm
	models.SaveTransactionForm
}

func (t *Transactions) saveTransaction(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	walletId := getWalletId(c)

	form := &saveTransactionForm{}
	if err := c.Bind(form); err != nil {
		return err
	}

	_, err := t.transactionService.SaveTransaction(c.Request().Context(), &form.SaveTransactionForm, walletId, user)
	if err != nil {
		var invalidForm *models.ErrInvalidForm
		if errors.As(err, &invalidForm) {
			saveError := HtmxEventSaveError{ErrorMessage: invalidForm.Message}
			saveErrorJson, err := json.Marshal(saveError)
			if err != nil {
				t.log.Error("Failed to marshal save error", "error", err)
				return err
			}

			c.Response().Header().Set(HtmxHeaderReswap, HtmxSwapNone)
			c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, string(saveErrorJson))

			return c.String(http.StatusOK, "save error")
		} else {
			return err
		}
	}

	fmt.Println(form.listTransactionsForm)
	fmt.Println(c.Request().URL.Query())

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventSaveSuccess)
	return t.renderTransactions(c, &form.listTransactionsForm, walletId, user)
}
