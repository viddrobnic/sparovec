package service

import (
	"context"
	"log/slog"

	"github.com/viddrobnic/sparovec/models"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	Update(ctx context.Context, transaction *models.Transaction) error
	List(ctx context.Context, req *models.TransactionsListRequest) ([]*models.Transaction, int, error)
	Delete(ctx context.Context, id int) error
}

type TransactionTagRepository interface {
	Get(ctx context.Context, tagId int) (*models.Tag, error)
	GetIds(ctx context.Context, tagIds []int) ([]*models.Tag, error)
}

type TransactionWalletRepository interface {
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
}

type Transaction struct {
	transactionRepository TransactionRepository
	tagRepository         TransactionTagRepository
	walletRepository      TransactionWalletRepository
	log                   *slog.Logger
}

func NewTransaction(transactionRepository TransactionRepository, tagRepository TransactionTagRepository, walletRepository TransactionWalletRepository, log *slog.Logger) *Transaction {
	return &Transaction{
		transactionRepository: transactionRepository,
		tagRepository:         tagRepository,
		walletRepository:      walletRepository,
		log:                   log,
	}
}

func (t *Transaction) Create(ctx context.Context, transaction *models.Transaction, user *models.User) error {
	hasPermission, err := t.walletRepository.HasPermission(ctx, transaction.WalletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	if err := t.validateTag(ctx, transaction.Tag, transaction.WalletId); err != nil {
		return err
	}

	if err := t.transactionRepository.Create(ctx, transaction); err != nil {
		t.log.ErrorContext(ctx, "Failed to create transaction", "error", err)
		return models.ErrInternalServer
	}

	return nil
}

func (t *Transaction) Update(ctx context.Context, transaction *models.Transaction, user *models.User) error {
	hasPermission, err := t.walletRepository.HasPermission(ctx, transaction.WalletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	if err := t.validateTag(ctx, transaction.Tag, transaction.WalletId); err != nil {
		return err
	}

	if err := t.transactionRepository.Update(ctx, transaction); err != nil {
		t.log.ErrorContext(ctx, "Failed to update transaction", "error", err)
		return models.ErrInternalServer
	}

	return nil
}

func (t *Transaction) validateTag(ctx context.Context, tag *models.Tag, walletId int) error {
	if tag == nil {
		return nil
	}

	tag, err := t.tagRepository.Get(ctx, tag.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tag", "error", err)
		return models.ErrInternalServer
	}

	if tag == nil {
		return &models.ErrInvalidForm{Message: "Invalid tag"}
	}

	if tag.WalletId != walletId {
		return &models.ErrInvalidForm{Message: "Invalid tag"}
	}

	return nil
}

func (t *Transaction) List(ctx context.Context, req *models.TransactionsListRequest, user *models.User) (*models.PaginatedResponse[*models.Transaction], error) {
	hasPermission, err := t.walletRepository.HasPermission(ctx, req.WalletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return &models.PaginatedResponse[*models.Transaction]{
			Count: 0,
			Data:  []*models.Transaction{},
		}, nil
	}

	transactions, count, err := t.transactionRepository.List(ctx, req)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to list transactions", "error", err)
		return nil, models.ErrInternalServer
	}

	if err := t.expandTags(ctx, transactions); err != nil {
		return nil, err
	}

	return &models.PaginatedResponse[*models.Transaction]{
		Count: count,
		Data:  transactions,
	}, nil
}

func (t *Transaction) Delete(ctx context.Context, id, walletId int, user *models.User) error {
	hasPermission, err := t.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	if err := t.transactionRepository.Delete(ctx, id); err != nil {
		t.log.ErrorContext(ctx, "Failed to delete transaction", "error", err)
		return models.ErrInternalServer
	}

	return nil
}

func (t *Transaction) expandTags(ctx context.Context, transactions []*models.Transaction) error {
	ids := make([]int, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.Tag != nil {
			ids = append(ids, transaction.Tag.Id)
		}
	}

	tags, err := t.tagRepository.GetIds(ctx, ids)
	if err != nil {
		t.log.ErrorContext(ctx, "Failed to get tags", "error", err)
		return models.ErrInternalServer
	}

	tagsMap := make(map[int]*models.Tag)
	for _, tag := range tags {
		tagsMap[tag.Id] = tag
	}

	for _, transaction := range transactions {
		if transaction.Tag != nil {
			transaction.Tag = tagsMap[transaction.Tag.Id]
		}
	}

	return nil
}
