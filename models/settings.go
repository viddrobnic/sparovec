package models

type Member struct {
	Id       int
	Username string
	IsSelf   bool
}

type SettingsContext struct {
	Navbar *NavbarContext

	WalletName string
	Members    []Member
}
