package models

import (
	"math"
	"strconv"
	"strings"
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
	Tag       *Tag
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
		Tag:       t.Tag,
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

type TransactionsListRequest struct {
	WalletId int
	Page     *Page
}

type TransactionsContext struct {
	Navbar *NavbarContext

	Transactions []*TransactionRender
	Tags         []*Tag
	CurrentPage  int
	PageSize     int
	TotalPages   int
}

type TransactionFormSubmitType string

const (
	TransactionFormSubmitTypeCreate TransactionFormSubmitType = "create"
	TransactionFormSubmitTypeEdit   TransactionFormSubmitType = "update"
)

type TransactionType string

const (
	TransactionTypeOutcome TransactionType = "outcome"
	TransactionTypeIncome  TransactionType = "income"
)

type SaveTransactionForm struct {
	Id         int                       `form:"id"`
	SubmitType TransactionFormSubmitType `form:"submit_type"`
	Name       string                    `form:"name"`
	Type       TransactionType           `form:"type"`
	Value      string                    `form:"value"`
	TagId      string                    `form:"tag"`
	Date       string                    `form:"date"`
}

func (f *SaveTransactionForm) Parse(walletId int) (*Transaction, error) {
	// Parse value
	valueStr := strings.ReplaceAll(f.Value, ",", ".")
	valueF, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, &ErrInvalidForm{Message: "Value is not a number"}
	}

	value := int(math.Round(valueF * 100))
	if value < 0 {
		return nil, &ErrInvalidForm{Message: "Value must be positive"}
	}

	if f.Type == TransactionTypeOutcome {
		value *= -1
	}

	// Parse date
	date, err := time.Parse("2006-01-02", f.Date)
	if err != nil {
		return nil, &ErrInvalidForm{Message: "Invalid date"}
	}

	// Parse tag
	var tag *Tag
	if f.TagId != "" {
		tagId, err := strconv.Atoi(f.TagId)
		if err != nil {
			return nil, &ErrInvalidForm{Message: "Invalid tag"}
		}

		tag = &Tag{Id: tagId}
	}

	return &Transaction{
		Id:        f.Id,
		WalletId:  walletId,
		Name:      f.Name,
		Value:     value,
		CreatedAt: date,
		Tag:       tag,
	}, nil
}
