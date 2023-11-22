package routes

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type Transactions struct {
	navbarService NavbarWalletsService
	log           *slog.Logger

	// Templates
	transactionsTemplate *template.Template
}

func NewTransactions(
	navbarService NavbarWalletsService,
	log *slog.Logger,
) *Transactions {
	transactionsTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/layout.html",
		"templates/transactions/transactions.html",
	))
	template.Must(transactionsTemplate.ParseGlob("templates/transactions/components/*"))

	return &Transactions{
		navbarService: navbarService,
		log:           log,

		transactionsTemplate: transactionsTemplate,
	}
}

func (t *Transactions) Mount(group *echo.Group) {
	group.Use(auth.RequiredMiddleware)

	group.GET("", t.transactions)

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
	}

	return t.transactionsTemplate.Execute(c.Response().Writer, ctx)
}
