package auth

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

func (r *RepositoryImpl) Insert(ctx context.Context, username, password, salt string) (*models.UserCredentials, error) {
	builder := sq.Insert("users").
		Columns("username", "password", "salt").
		Values(username, password, salt).
		Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.UserCredentials{}
	err = r.db.GetContext(ctx, user, stmt, args...)
	return user, err
}

func (r *RepositoryImpl) GetByUsername(ctx context.Context, username string) (*models.UserCredentials, error) {
	builder := sq.Select("*").From("users").Where("username = ?", username)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.UserCredentials{}
	err = r.db.GetContext(ctx, user, stmt, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}
