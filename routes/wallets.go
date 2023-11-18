package routes

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viddrobnic/sparovec/models"
)

type Wallets struct {
	walletsTemplate *template.Template
}

func NewWallets() *Wallets {
	walletsTemplate := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/layout.html",
		"templates/wallets/wallets.html",
	))

	return &Wallets{
		walletsTemplate: walletsTemplate,
	}
}

func (w *Wallets) Mount(group *echo.Group) {
	group.GET("/", w.wallets)
}

func (w *Wallets) wallets(c echo.Context) error {
	user, _ := c.Get(models.UserContextKey).(*models.User)

	ctx := &models.WalletsContext{
		Navbar: &models.NavbarContext{
			SelectedWallet: 0,
			Wallets: []models.NavbarWallet{
				{
					Id:       1,
					Name:     "Test wallet 1",
					Selected: false,
				},
				{
					Id:       2,
					Name:     "Test wallet 2",
					Selected: false,
				},
			},
			Username: user.Username,
		},
	}

	c.Response().WriteHeader(http.StatusOK)
	return w.walletsTemplate.Execute(c.Response().Writer, ctx)
}
