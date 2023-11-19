package models

import "time"

type WalletsContext struct {
	Navbar *NavbarContext

	Wallets []*Wallet
}

type Wallet struct {
	Id        int
	Name      string
	CreatedAt time.Time `db:"created_at"`
}
