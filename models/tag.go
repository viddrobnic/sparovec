package models

import "time"

type Tag struct {
	Id        int
	WalletId  int `db:"wallet_id"`
	Name      string
	CreatedAt time.Time `db:"created_at"`
}

type TagsContext struct {
	Navbar *NavbarContext

	Tags []*Tag
}
