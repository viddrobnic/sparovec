package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/models"
)

type Wallets struct {
	db *sqlx.DB
}

func NewWallets(db *sqlx.DB) *Wallets {
	return &Wallets{db: db}
}

func (w *Wallets) ForUser(ctx context.Context, userId int) ([]*models.Wallet, error) {
	builder := sq.Select("w.*").
		From("wallets w").
		Join("wallet_users wu ON w.id = wu.wallet_id").
		Where("wu.user_id = ?", userId).
		OrderBy("w.created_at desc", "w.id")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	wallets := []*models.Wallet{}
	err = w.db.SelectContext(ctx, &wallets, stmt, args...)
	return wallets, err
}

func (w *Wallets) Create(ctx context.Context, userId int, name string) (*models.Wallet, error) {
	tx, err := w.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Insert wallet
	builder := sq.Insert("wallets").Columns("name").Values(name).Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	wallet := &models.Wallet{}
	err = tx.GetContext(ctx, wallet, stmt, args...)
	if err != nil {
		return nil, err
	}

	// Insert wallet user
	builder = sq.Insert("wallet_users").Columns("user_id", "wallet_id").Values(userId, wallet.Id)

	stmt, args, err = builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	return wallet, err
}
