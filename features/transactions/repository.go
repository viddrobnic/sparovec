package transactions

import (
	"context"
	"database/sql"
	"fmt"

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

func (t *RepositoryImpl) CreateMany(ctx context.Context, transactions []*models.Transaction) error {
	builder := sq.Insert("transactions").
		Columns(
			"wallet_id",
			"name",
			"value",
			"tag_id",
			"created_at",
		)

	for _, tr := range transactions {
		var tagId *int
		if tr.Tag != nil {
			tagId = &tr.Tag.Id
		}

		builder = builder.Values(
			tr.WalletId,
			tr.Name,
			tr.Value,
			tagId,
			tr.CreatedAt,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("query toSql: %w", err)
	}

	_, err = t.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

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

func (t *RepositoryImpl) TagInfoForNames(ctx context.Context, walletId int, names []string) (map[string]int, error) {
	innerBuilder := sq.StatementBuilder.
		Select("*").
		Prefix("NOT EXISTS(").
		From("transactions tr_inner").
		Where("tr_inner.tag_id IS NOT NULL").
		Where("tr_inner.wallet_id = tr.wallet_id").
		Where("tr_inner.name = tr.name").
		Where("tr_inner.created_at > tr.created_at").
		Suffix(")")

	builder := sq.
		Select("*").
		From("transactions tr").
		Where(sq.Eq{
			"wallet_id": walletId,
			"name":      names,
		}).
		Where("tag_id is not null").
		Where(innerBuilder).
		OrderBy("created_at DESC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder toSql: %w", err)
	}

	transactions := []*models.DbTransaction{}
	err = t.db.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}

	res := make(map[string]int)
	for _, tr := range transactions {
		if tr.TagId.Valid {
			res[tr.Name] = int(tr.TagId.Int32)
		}
	}

	return res, nil
}
