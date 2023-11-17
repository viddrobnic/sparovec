package repository

import (
	"context"

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

func (u *Users) Insert(ctx context.Context, username, password, salt string) (*models.User, error) {
	builder := sq.Insert("users").
		Columns("username", "password", "salt").
		Values(username, password, salt).
		Suffix("RETURNING *")

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	err = u.db.GetContext(ctx, user, sql, args...)
	return user, err
}
