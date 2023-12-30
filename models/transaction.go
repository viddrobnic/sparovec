package models

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
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
	Id            int
	Name          string
	Value         string
	FormValue     string
	Type          string
	Tag           *Tag
	FormTagId     string
	CreatedAt     string
	FormCreatedAt string
}

func (t *Transaction) Render() *TransactionRender {
	p := message.NewPrinter(language.Slovenian)
	value := p.Sprintf("%.2f", float64(t.Value)/100)
	value += " €"

	var transactionType string
	if t.Value < 0 {
		transactionType = "outcome"
	} else {
		transactionType = "income"
	}

	formTagId := ""
	if t.Tag != nil {
		formTagId = strconv.Itoa(t.Tag.Id)
	}

	return &TransactionRender{
		Id:            t.Id,
		Name:          t.Name,
		Value:         value,
		FormValue:     fmt.Sprintf("%.2f", math.Abs(float64(t.Value))/100),
		Type:          transactionType,
		Tag:           t.Tag,
		FormTagId:     formTagId,
		CreatedAt:     t.CreatedAt.Format("02. 01. 2006"),
		FormCreatedAt: t.CreatedAt.Format("2006-01-02"),
	}
}

func RenderTransactions(transactions []*Transaction) []*TransactionRender {
	rendered := make([]*TransactionRender, len(transactions))

	for i, transaction := range transactions {
		rendered[i] = transaction.Render()
	}

	return rendered
}

type TransactionsListRequest struct {
	WalletId int
	Page     *Page
}

type TransactionsContext struct {
	Navbar *NavbarContext

	Transactions []*TransactionRender
	Tags         []*Tag
	CurrentPage  int
	TotalPages   int

	UrlParams       string
	PreviousPageUrl string
	NextPageUrl     string
}

type DbTransaction struct {
	Id        int       `db:"id"`
	WalletId  int       `db:"wallet_id"`
	Name      string    `db:"name"`
	Value     int       `db:"value"`
	CreatedAt time.Time `db:"created_at"`

	TagId sql.NullInt32 `db:"tag_id"`
}

func (dt *DbTransaction) ToModel() *Transaction {
	var tag *Tag
	if dt.TagId.Valid {
		tag = &Tag{
			Id: int(dt.TagId.Int32),
		}
	}

	return &Transaction{
		Id:        dt.Id,
		WalletId:  dt.WalletId,
		Name:      dt.Name,
		Value:     dt.Value,
		Tag:       tag,
		CreatedAt: dt.CreatedAt,
	}
}
