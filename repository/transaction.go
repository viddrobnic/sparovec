package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/models"
)

type dbTransaction struct {
	Id        int       `db:"id"`
	WalletId  int       `db:"wallet_id"`
	Name      string    `db:"name"`
	Value     int       `db:"value"`
	CreatedAt time.Time `db:"created_at"`

	TagId sql.NullInt32 `db:"tag_id"`
}

func (dt *dbTransaction) ToModel() *models.Transaction {
	var tag *models.Tag
	if dt.TagId.Valid {
		tag = &models.Tag{
			Id: int(dt.TagId.Int32),
		}
	}

	return &models.Transaction{
		Id:        dt.Id,
		WalletId:  dt.WalletId,
		Name:      dt.Name,
		Value:     dt.Value,
		Tag:       tag,
		CreatedAt: dt.CreatedAt,
	}
}

type Transaction struct {
	db *sqlx.DB
}

func NewTransaction(db *sqlx.DB) *Transaction {
	return &Transaction{db: db}
}

func (t *Transaction) Create(ctx context.Context, transaction *models.Transaction) error {
	var tagId *int
	if transaction.Tag != nil {
		tagId = &transaction.Tag.Id
	}

	builder := sq.Insert("transactions").
		Columns(
			"wallet_id",
			"name",
			"value",
			"tag_id",
			"created_at",
		).Values(
		transaction.WalletId,
		transaction.Name,
		transaction.Value,
		tagId,
		transaction.CreatedAt,
	).Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	dbTransaction := &dbTransaction{}
	err = t.db.GetContext(ctx, dbTransaction, stmt, args...)
	if err != nil {
		return err
	}

	*transaction = *dbTransaction.ToModel()
	return nil
}
