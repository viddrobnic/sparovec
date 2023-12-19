package service

import (
	"context"
	"log/slog"

	"github.com/viddrobnic/sparovec/models"
)

type SettingsWalletRepository interface {
	HasPermission(ctx context.Context, walletId, userId int) (bool, error)
	ForId(ctx context.Context, walletId int) (*models.Wallet, error)
	Members(ctx context.Context, walletId int) ([]*models.Member, error)
	SetName(ctx context.Context, walletId int, name string) error
	AddMember(ctx context.Context, walletId, userId int) error
	RemoveMember(ctx context.Context, walletId int, userId string) error
	Delete(ctx context.Context, walletId int) error
}

type SettingsUserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.UserCredentials, error)
}

type Settings struct {
	walletRepository SettingsWalletRepository
	userRepository   SettingsUserRepository

	log *slog.Logger
}

func NewSettings(
	walletRepository SettingsWalletRepository,
	userRepository SettingsUserRepository,
	log *slog.Logger,
) *Settings {
	return &Settings{
		walletRepository: walletRepository,
		userRepository:   userRepository,

		log: log,
	}
}

func (s *Settings) WalletName(ctx context.Context, walletId int, user *models.User) (string, error) {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return "", models.ErrInternalServer
	}

	if !hasPermission {
		return "", nil
	}

	wall, err := s.walletRepository.ForId(ctx, walletId)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet", "error", err)
		return "", models.ErrInternalServer
	}

	if wall == nil {
		return "", nil
	}

	return wall.Name, nil
}

func (s *Settings) Members(ctx context.Context, walletId int, user *models.User) ([]*models.Member, error) {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return nil, models.ErrInternalServer
	}

	if !hasPermission {
		return nil, nil
	}

	members, err := s.walletRepository.Members(ctx, walletId)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet members", "error", err)
		return nil, models.ErrInternalServer
	}

	sortedMembers := make([]*models.Member, 1, len(members))
	for _, member := range members {
		if member.Id == user.Id {
			member.IsSelf = true
			sortedMembers[0] = member
		} else {
			sortedMembers = append(sortedMembers, member)
		}
	}

	return sortedMembers, nil
}

func (s *Settings) ChangeWalletName(ctx context.Context, walletId int, name string, user *models.User) error {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	return s.walletRepository.SetName(ctx, walletId, name)
}

func (s *Settings) AddMember(ctx context.Context, walletId int, username string, user *models.User) error {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	creds, err := s.userRepository.GetByUsername(ctx, username)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get user by username", "error", err)
		return models.ErrInternalServer
	}

	if creds == nil {
		return nil
	}

	return s.walletRepository.AddMember(ctx, walletId, creds.Id)
}

func (s *Settings) RemoveMember(ctx context.Context, walletId int, id string, user *models.User) error {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	return s.walletRepository.RemoveMember(ctx, walletId, id)
}

func (s *Settings) DeleteWallet(ctx context.Context, walletId int, user *models.User) error {
	hasPermission, err := s.walletRepository.HasPermission(ctx, walletId, user.Id)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to get wallet permission", "error", err)
		return models.ErrInternalServer
	}

	if !hasPermission {
		return nil
	}

	return s.walletRepository.Delete(ctx, walletId)
}
