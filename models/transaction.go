package models

import (
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Transaction struct {
	Id        int
	WalletId  int
	Name      string
	Value     int
	Tag       *Tag
	CreatedAt time.Time
}
type TransactionRender struct {
	Id        int
	Name      string
	Value     string
	TagName   string
	CreatedAt string
}

func (t *Transaction) Render() *TransactionRender {
	p := message.NewPrinter(language.Slovenian)
	value := p.Sprintf("%.2f", float64(t.Value)/100)
	value += " €"

	return &TransactionRender{
		Id:        t.Id,
		Name:      t.Name,
		Value:     value,
		TagName:   t.Tag.Name,
		CreatedAt: t.CreatedAt.Format("02. 01. 2006"),
	}
}

func RenderTransactions(transactions []*Transaction) []*TransactionRender {
	rendered := make([]*TransactionRender, len(transactions))

	for i, transaction := range transactions {
		rendered[i] = transaction.Render()
	}

	return rendered
}

type TransactionsContext struct {
	Navbar *NavbarContext

	Transactions []*TransactionRender
}
