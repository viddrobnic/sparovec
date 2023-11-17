package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"

	"github.com/viddrobnic/sparovec/models"
	"golang.org/x/crypto/argon2"
)

const saltLenght = 16

type UserRepository interface {
	Insert(ctx context.Context, username, password, salt string) (*models.User, error)
}

type User struct {
	repository UserRepository
	log        *slog.Logger
}

func NewUser(repository UserRepository, log *slog.Logger) *User {
	return &User{
		repository: repository,
		log:        log,
	}
}

func (u *User) Create(ctx context.Context, username, password string) (*models.User, error) {
	// Create saltBytes
	saltBytes := make([]byte, saltLenght)
	_, err := rand.Read(saltBytes)
	if err != nil {
		u.log.Error("Failed to generate salt", "error", err)
		return nil, models.ErrInternalServer
	}

	salt := base64.StdEncoding.EncodeToString(saltBytes)

	// Hash password
	hashedPasswordBytes := hashPassword([]byte(password), saltBytes)
	hashedPassword := base64.StdEncoding.EncodeToString(hashedPasswordBytes)

	// Insert user
	user, err := u.repository.Insert(ctx, username, hashedPassword, salt)
	if err != nil {
		u.log.Error("Failed to insert user", "error", err)
		return nil, models.ErrInternalServer
	}

	return user, nil
}

func hashPassword(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}
