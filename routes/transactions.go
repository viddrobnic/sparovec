package routes

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type TransactionService interface {
	SaveTransaction(ctx context.Context, transaction *models.TransactionForm, walletId int) (*models.Transaction, error)
}

type Transactions struct {
	navbarService      NavbarWalletsService
	transactionService TransactionService
	log                *slog.Logger

	// Templates
	transactionsTemplate *template.Template
}

func NewTransactions(
	navbarService NavbarWalletsService,
	transactionService TransactionService,
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

func (t *Transactions) transactions(c echo.Context) error {
	navbarCtx, err := createNavbarContext(c, t.navbarService)
	if err != nil {
		t.log.ErrorContext(c.Request().Context(), "Failed to create navbar context", "error", err)
		return err
	}

	ctx := &models.TransactionsContext{
		Navbar:       navbarCtx,
		Transactions: []*models.TransactionRender{},
		Tags:         []*models.Tag{},
	}

	return t.transactionsTemplate.Execute(c.Response().Writer, ctx)
}

func (t *Transactions) saveTransaction(c echo.Context) error {
	walletId := getWalletId(c)

	form := &models.TransactionForm{}
	if err := c.Bind(form); err != nil {
		return err
	}

	_, err := t.transactionService.SaveTransaction(c.Request().Context(), form, walletId)
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

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventSaveSuccess)
	return t.transactions(c)
}
