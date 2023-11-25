package routes

import (
	"context"
	"net/http"

	"github.com/viddrobnic/sparovec/middleware/auth"
	"github.com/viddrobnic/sparovec/models"
)

type NavbarWalletsService interface {
	ForUser(ctx context.Context, userId int) ([]*models.Wallet, error)
}

func createNavbarContext(r *http.Request, walletsService NavbarWalletsService) (*models.NavbarContext, error) {
	user := auth.GetUser(r)

	walletId := getWalletId(r)
	wallets, err := walletsService.ForUser(r.Context(), user.Id)
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
		SelectedWalletId: walletId,
		Wallets:          navbarWallets,
		Username:         user.Username,
	}, nil
}
