package models

type NavbarWallet struct {
	Id       int
	Name     string
	Selected bool
}

type NavbarContext struct {
	SelectedWalletId int
	Wallets          []NavbarWallet

	Username string
}

type Navbar struct {
	SelectedWalletId int
	Wallets          []*Wallet
	Username         string
	Title            string
}
