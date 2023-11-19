package routes

import (
	"context"
	"html/template"
	"log/slog"

	"github.com/labstack/echo/v4"
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

func NewWallets(service WalletsService, log *slog.Logger) *Wallets {
	walletsTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/layout.html",
		"templates/wallets/wallets.html",
	))
	template.Must(walletsTemplate.ParseGlob("templates/wallets/components/*"))

	walletCardTemplate := template.Must(template.ParseFiles("templates/wallets/components/wallet-card.html"))

	return &Wallets{
		service: service,
		log:     log,

		walletsTemplate:    walletsTemplate,
		walletCardTemplate: walletCardTemplate,
	}
}

func (w *Wallets) Mount(group *echo.Group) {
	group.Use(auth.RequiredMiddleware)

	group.GET("/", w.wallets)
	group.POST("/", w.createWallet)
}

func (w *Wallets) wallets(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	navbarCtx, err := createNavbarContext(c, w.service)
	if err != nil {
		w.log.Error("Failed to create navbar context", "error", err)
		return err
	}

	wallets, err := w.service.ForUser(c.Request().Context(), user.Id)
	if err != nil {
		w.log.Error("Failed to get wallets", "error", err)
		return err
	}

	ctx := &models.WalletsContext{
		Navbar:  navbarCtx,
		Wallets: wallets,
	}

	return renderTemplate(c, w.walletsTemplate, ctx)
}

func (w Wallets) createWallet(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)
	name := c.FormValue("name")

	wallet, err := w.service.Create(c.Request().Context(), user.Id, name)
	if err != nil {
		w.log.Error("Failed to create wallet", "error", err)

		// TODO: Better error handling
		return err
	}

	c.Response().Header().Set(HtmxHeaderTriggerAfterSettle, HtmxEventCreateSuccess)
	return renderTemplateNamed(c, w.walletCardTemplate, "wallet-card", wallet)
}
