package transactions

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/viddrobnic/sparovec/models"
)

type listTransactionsForm struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func listTransactionFormFromRequest(r *http.Request) *listTransactionsForm {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	return &listTransactionsForm{
		Page:     page,
		PageSize: pageSize,
	}
}

type transactionFormSubmitType string

const (
	transactionFormSubmitTypeCreate transactionFormSubmitType = "create"
	transactionFormSubmitTypeEdit   transactionFormSubmitType = "update"
)

type transactionType string

const (
	transactionTypeOutcome transactionType = "outcome"
	transactionTypeIncome  transactionType = "income"
)

type saveTransactionForm struct {
	Id         int                       `form:"id"`
	SubmitType transactionFormSubmitType `form:"submit_type"`
	Name       string                    `form:"name"`
	Type       transactionType           `form:"type"`
	Value      string                    `form:"value"`
	TagId      string                    `form:"tag"`
	Date       string                    `form:"date"`
}

func parseTransactionValue(valueStr string) (int, error) {
	valueStr = strings.ReplaceAll(valueStr, ",", ".")
	valueF, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, &models.ErrInvalidForm{Message: "Value is not a number"}
	}

	return int(math.Round(valueF * 100)), nil
}

func (f *saveTransactionForm) parse(walletId int) (*models.Transaction, error) {
	// Parse value
	value, err := parseTransactionValue(f.Value)
	if err != nil {
		return nil, err
	}

	if value < 0 {
		return nil, &models.ErrInvalidForm{Message: "Value must be positive"}
	}

	if f.Type == transactionTypeOutcome {
		value *= -1
	}

	// Parse date
	date, err := time.Parse("2006-01-02", f.Date)
	if err != nil {
		return nil, &models.ErrInvalidForm{Message: "Invalid date"}
	}

	// Parse tag
	var tag *models.Tag
	if f.TagId != "" {
		tagId, err := strconv.Atoi(f.TagId)
		if err != nil {
			return nil, &models.ErrInvalidForm{Message: "Invalid tag"}
		}

		tag = &models.Tag{Id: tagId}
	}

	return &models.Transaction{
		Id:        f.Id,
		WalletId:  walletId,
		Name:      f.Name,
		Value:     value,
		CreatedAt: date,
		Tag:       tag,
	}, nil
}

func saveTransactionFormFromRequest(r *http.Request) *saveTransactionForm {
	id, _ := strconv.Atoi(r.FormValue("id"))
	submitType := transactionFormSubmitType(r.FormValue("submit_type"))
	transactionType := transactionType(r.FormValue("type"))

	return &saveTransactionForm{
		Id:         id,
		SubmitType: submitType,
		Name:       r.FormValue("name"),
		Type:       transactionType,
		Value:      r.FormValue("value"),
		TagId:      r.FormValue("tag"),
		Date:       r.FormValue("date"),
	}
}
