package tags

import (
	"context"

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

func (t *RepositoryImpl) List(ctx context.Context, walletId int) ([]*models.Tag, error) {
	builder := sq.Select("*").
		From("tags").
		Where("wallet_id = ?", walletId).
		OrderBy("name", "id")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tags := []*models.Tag{}
	err = t.db.SelectContext(ctx, &tags, stmt, args...)
	return tags, err
}

func (t *RepositoryImpl) Create(ctx context.Context, walletId int, name string) (*models.Tag, error) {
	builder := sq.Insert("tags").
		Columns("wallet_id", "name").
		Values(walletId, name).
		Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tag := &models.Tag{}
	err = t.db.GetContext(ctx, tag, stmt, args...)
	return tag, err
}

func (t *RepositoryImpl) Get(ctx context.Context, tagId int) (*models.Tag, error) {
	builder := sq.Select("*").
		From("tags").
		Where("id = ?", tagId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tag := &models.Tag{}
	err = t.db.GetContext(ctx, tag, stmt, args...)
	return tag, err
}

func (t *RepositoryImpl) Update(ctx context.Context, tagId int, name string) (*models.Tag, error) {
	builder := sq.Update("tags").
		Set("name", name).
		Where("id = ?", tagId).
		Suffix("RETURNING *")

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tag := &models.Tag{}
	err = t.db.GetContext(ctx, tag, stmt, args...)
	return tag, err
}

func (t *RepositoryImpl) Delete(ctx context.Context, tagId int) error {
	builder := sq.Delete("tags").Where("id = ?", tagId)

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = t.db.ExecContext(ctx, stmt, args...)
	return err
}

func (t *RepositoryImpl) GetIds(ctx context.Context, ids []int) ([]*models.Tag, error) {
	if len(ids) == 0 {
		return []*models.Tag{}, nil
	}

	builder := sq.Select("*").
		From("tags").
		Where(sq.Eq{"id": ids})

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tags := []*models.Tag{}
	err = t.db.SelectContext(ctx, &tags, stmt, args...)
	return tags, err
}
