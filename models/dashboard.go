package models

type TagBalance struct {
	Tag     *Tag
	Balance int
}

type DashboardData struct {
	Income         int
	Outcome        int
	Balance        int
	NrTransactions int
	TagBalance     []TagBalance
}
