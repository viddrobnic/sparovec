package dashboard

import (
	"context"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/models"
)

type RepositoryImpl struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) GetTransactions(ctx context.Context, walletId, year int, month time.Month) ([]*models.Transaction, error) {
	builder := sq.Select("*").
		From("transactions").
		Where(sq.Eq{
			"wallet_id":                  walletId,
			"STRFTIME('%Y', created_at)": strconv.Itoa(year),
			"STRFTIME('%m', created_at)": strconv.Itoa(int(month)),
		})

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	dbTransactions := []*models.DbTransaction{}
	err = r.db.SelectContext(ctx, &dbTransactions, stmt, args...)
	if err != nil {
		return nil, err
	}

	transactions := make([]*models.Transaction, len(dbTransactions))
	for i, dbTransaction := range dbTransactions {
		transactions[i] = dbTransaction.ToModel()
	}

	return transactions, nil
}
