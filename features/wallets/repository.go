package wallets

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/models"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (w *Repository) ForUser(ctx context.Context, userId int) ([]*models.Wallet, error) {
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

func (w *Repository) ForId(ctx context.Context, walletId int) (*models.Wallet, error) {
	builder := sq.Select("*").From("wallets").Where("id = ?", walletId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	wallet := &models.Wallet{}
	err = w.db.GetContext(ctx, wallet, stmt, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return wallet, err
}

func (w *Repository) SetName(ctx context.Context, walletId int, name string) error {
	builder := sq.Update("wallets").Set("name", name).Where("id = ?", walletId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, stmt, args...)
	return err
}

func (w *Repository) Members(ctx context.Context, walletId int) ([]*models.Member, error) {
	builder := sq.Select("u.id", "u.username").
		From("users u").
		InnerJoin("wallet_users wu ON u.id = wu.user_id").
		Where("wu.wallet_id = ?", walletId).
		OrderBy("u.username")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	members := []*models.Member{}
	err = w.db.SelectContext(ctx, &members, stmt, args...)
	return members, err
}

func (w *Repository) AddMember(ctx context.Context, walletId, userId int) error {
	builder := sq.Insert("wallet_users").
		Columns("wallet_id", "user_id").
		Values(walletId, userId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, stmt, args...)
	return err
}

func (w *Repository) RemoveMember(ctx context.Context, walletId int, userId string) error {
	builder := sq.Delete("wallet_users").
		Where(sq.Eq{
			"wallet_id": walletId,
			"user_id":   userId,
		})

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, stmt, args...)
	return err
}

func (w *Repository) HasPermission(ctx context.Context, walletId, userId int) (bool, error) {
	builder := sq.Select("1").
		From("wallet_users").
		Where(sq.Eq{
			"wallet_id": walletId,
			"user_id":   userId,
		}).Limit(1)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return false, err
	}

	var exists bool
	err = w.db.GetContext(ctx, &exists, stmt, args...)
	if err == sql.ErrNoRows {
		exists = false
	} else if err != nil {
		return false, err
	}

	return exists, nil
}

func (w *Repository) Create(ctx context.Context, userId int, name string) (*models.Wallet, error) {
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

func (w *Repository) Delete(ctx context.Context, walletId int) error {
	builder := sq.Delete("wallets").Where("id = ?", walletId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, stmt, args...)
	return err
}
