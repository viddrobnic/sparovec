package service

import (
	"context"
	"log/slog"

	"github.com/viddrobnic/sparovec/models"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
}

type TransactionTagRepository interface {
	Get(ctx context.Context, tagId int) (*models.Tag, error)
}

type Transaction struct {
	transactionRepository TransactionRepository
	tagRepository         TransactionTagRepository
	log                   *slog.Logger
}

func NewTransaction(transactionRepository TransactionRepository, tagRepository TransactionTagRepository, log *slog.Logger) *Transaction {
	return &Transaction{
		transactionRepository: transactionRepository,
		tagRepository:         tagRepository,
		log:                   log,
	}
}

func (t *Transaction) SaveTransaction(ctx context.Context, form *models.TransactionForm, walletId int) (*models.Transaction, error) {
	transaction, err := form.Parse(walletId)
	if err != nil {
		return nil, err
	}

	switch form.SubmitType {
	case models.TransactionFormSubmitTypeCreate:
		return transaction, t.createTransaction(ctx, transaction)
	case models.TransactionFormSubmitTypeEdit:
		return transaction, t.updateTransaction(ctx, transaction)
	default:
		return nil, &models.ErrInvalidForm{Message: "Invalid submit type"}
	}
}

func (t *Transaction) createTransaction(ctx context.Context, transaction *models.Transaction) error {
	if err := t.validateTag(ctx, transaction.Tag, transaction.WalletId); err != nil {
		return err
	}

	if err := t.transactionRepository.Create(ctx, transaction); err != nil {
		t.log.ErrorContext(ctx, "Failed to create transaction", "error", err)
		return models.ErrInternalServer
	}

	return nil
}

func (t *Transaction) updateTransaction(ctx context.Context, transaction *models.Transaction) error {
	if err := t.validateTag(ctx, transaction.Tag, transaction.WalletId); err != nil {
		return err
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
