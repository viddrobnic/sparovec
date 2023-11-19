package routes

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/models"
)

type NavbarWalletsService interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
}

func createNavbarContext(c echo.Context, walletsService NavbarWalletsService) (*models.NavbarContext, error) {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	walletId, _ := strconv.Atoi(c.Param("walletId"))

	wallets, err := walletsService.ForUser(c.Request().Context(), user.Id)
	if err != nil {
		return nil, err
	}

	navbarWallets := make([]models.NavbarWallet, len(wallets))
	for i, wallet := range wallets {
		navbarWallets[i] = models.NavbarWallet{
			Id:       wallet.Id,
			Name:     wallet.Name,
			Selected: wallet.Id == walletId,
		}
	}

	return &models.NavbarContext{
		SelectedWallet: walletId,
		Wallets:        navbarWallets,
		Username:       user.Username,
	}, nil
}
