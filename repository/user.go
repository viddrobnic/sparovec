package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/viddrobnic/sparovec/models"
)

type Users struct {
	db *sqlx.DB
}

func NewUsers(db *sqlx.DB) *Users {
	return &Users{db: db}
}

func (u *Users) Insert(ctx context.Context, username, password, salt string) (*models.UserCredentials, error) {
	builder := sq.Insert("users").
		Columns("username", "password", "salt").
		Values(username, password, salt).
		Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.UserCredentials{}
	err = u.db.GetContext(ctx, user, stmt, args...)
	return user, err
}

func (u *Users) GetByUsername(ctx context.Context, username string) (*models.UserCredentials, error) {
	builder := sq.Select("*").From("users").Where("username = ?", username)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.UserCredentials{}
	err = u.db.GetContext(ctx, user, stmt, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}
