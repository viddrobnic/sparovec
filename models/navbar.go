package models

type NavbarWallet struct {
	Id       int
	Name     string
	Selected bool
}

type NavbarContext struct {
	SelectedWallet int
	Wallets        []NavbarWallet

	Username string
}
