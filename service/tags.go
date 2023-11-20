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
	tag, err := t.tagsRepository.Get(ctx, tagId)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tag", "error", err)
		return nil, models.ErrInternalServer
	}

	hasPermission, err := t.walletRepository.HasPermission(ctx, tag.WalletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return nil, models.ErrForbidden
	}

	tag, err = t.tagsRepository.Update(ctx, tagId, name)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to update tag", "error", err)
		return nil, models.ErrInternalServer
	}

	return tag, nil
}
