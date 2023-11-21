package service

import (
	"context"
	"log/slog"

	"github.com/viddrobnic/sparovec/models"
)

type TagsRepository interface {
	List(ctx context.Context, walletId int) ([]*models.Tag, error)
	Create(ctx context.Context, walletId int, name string) (*models.Tag, error)
	Get(ctx context.Context, tagId int) (*models.Tag, error)
	Update(ctx context.Context, tagId int, name string) (*models.Tag, error)
	Delete(ctx context.Context, tagId int) error
}

type TagsWalletRepository interface {
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
}

type Tags struct {
	tagsRepository   TagsRepository
	walletRepository TagsWalletRepository

	log *slog.Logger
}

func NewTags(
	tagsRepository TagsRepository,
	walletRepository TagsWalletRepository,
	log *slog.Logger,
) *Tags {
	return &Tags{
		tagsRepository:   tagsRepository,
		walletRepository: walletRepository,
		log:              log,
	}
}

func (t *Tags) List(ctx context.Context, walletId int, user *models.User) ([]*models.Tag, error) {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return []*models.Tag{}, nil
	}

	return t.tagsRepository.List(ctx, walletId)
}

func (t *Tags) Create(ctx context.Context, walletId int, name string, user *models.User) (*models.Tag, error) {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return nil, models.ErrForbidden
	}

	tag, err := t.tagsRepository.Create(ctx, walletId, name)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to create tag", "error", err)
		return nil, models.ErrInternalServer
	}

	return tag, nil
}

func (t *Tags) Update(ctx context.Context, tagId int, name string, user *models.User) (*models.Tag, error) {
	hasPermission, err := t.hasPermission(ctx, tagId, user)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, models.ErrForbidden
	}

	tag, err := t.tagsRepository.Update(ctx, tagId, name)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to update tag", "error", err)
		return nil, models.ErrInternalServer
	}

	return tag, nil
}

func (t *Tags) Delete(ctx context.Context, tagId int, user *models.User) error {
	hasPermission, err := t.hasPermission(ctx, tagId, user)
	if err != nil {
		return err
	}

	if !hasPermission {
		return models.ErrForbidden
	}

	err = t.tagsRepository.Delete(ctx, tagId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to delete tag", "error", err)
		return models.ErrInternalServer
	}

	return nil
}

func (t *Tags) hasPermission(ctx context.Context, tagId int, user *models.User) (bool, error) {
	tag, err := t.tagsRepository.Get(ctx, tagId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tag", "error", err)
		return false, models.ErrInternalServer
	}

	hasPermission, err := t.walletRepository.HasPermission(ctx, tag.WalletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return false, models.ErrInternalServer
	}

	return hasPermission, nil
}
