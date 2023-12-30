package transactions

import (
	"context"
	"database/sql"

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

func (t *RepositoryImpl) Create(ctx context.Context, transaction *models.Transaction) error {
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

	dbTransaction := &models.DbTransaction{}
	err = t.db.GetContext(ctx, dbTransaction, stmt, args...)
	if err != nil {
		return err
	}

	*transaction = *dbTransaction.ToModel()
	return nil
}

func (t *RepositoryImpl) Update(ctx context.Context, transaction *models.Transaction) error {
	var tagId sql.NullInt32
	if transaction.Tag != nil {
		tagId = sql.NullInt32{
			Int32: int32(transaction.Tag.Id),
			Valid: true,
		}
	}

	builder := sq.Update("transactions").
		Set("name", transaction.Name).
		Set("value", transaction.Value).
		Set("tag_id", tagId).
		Set("created_at", transaction.CreatedAt).
		Where("id = ?", transaction.Id)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = t.db.ExecContext(ctx, stmt, args...)
	return err
}

func (t *RepositoryImpl) List(ctx context.Context, req *models.TransactionsListRequest) ([]*models.Transaction, int, error) {
	builder := sq.Select("*").From("transactions").
		Where("wallet_id = ? ", req.WalletId)

	countBuilder := sq.Select("COUNT(*)").FromSelect(builder, "transactions")

	builder = builder.
		OrderBy("date(created_at) DESC", "value ASC", "name", "id").
		Offset(uint64(req.Page.Offset())).
		Limit(uint64(req.Page.Limit()))

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	dbTransactions := []*models.DbTransaction{}
	err = t.db.SelectContext(ctx, &dbTransactions, stmt, args...)
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*models.Transaction, len(dbTransactions))
	for i, dbTransaction := range dbTransactions {
		transactions[i] = dbTransaction.ToModel()
	}

	countStmt, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var count int
	err = t.db.GetContext(ctx, &count, countStmt, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

func (t *RepositoryImpl) Delete(ctx context.Context, id int) error {
	builder := sq.Delete("transactions").Where("id = ?", id)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = t.db.ExecContext(ctx, stmt, args...)
	return err
}
