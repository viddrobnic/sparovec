package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"github.com/viddrobnic/sparovec/models"
)

const saltLenght = 16

func (a *Auth) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	user, err := a.repository.GetByUsername(ctx, username)
	if err != nil {
		a.log.Error("Failed to get user", "error", err)
		return nil, models.ErrInternalServer
	}

	if user == nil {
		a.log.Info("User not found", "username", username)
		return nil, models.ErrInvalidCredentials
	}

	if !doPasswordsMatch(user.Password, user.Salt, password) {
		return nil, models.ErrInvalidCredentials
	}

	return &user.User, nil
}

func (a *Auth) CreateSession(user *models.User) (*models.Session, error) {
	// Create session
	expiresAt := time.Now().Add(time.Duration(a.conf.Auth.SessionTtl) * time.Second)
	sess := &models.Session{
		User:      user,
		ExpiresAt: expiresAt,
	}

	signatureBytes, err := signSession(sess, a.conf.Auth.SigningKey)
	if err != nil {
		a.log.Error("Failed to sign session", "error", err)
		return nil, models.ErrInternalServer
	}

	sess.Signature = base64.StdEncoding.EncodeToString(signatureBytes)

	return sess, nil
}

func (a *Auth) ValidateSession(session *models.Session) error {
	if session.ExpiresAt.Before(time.Now()) {
		return models.ErrInvalidCredentials
	}

	signatureBytes, err := signSession(session, a.conf.Auth.SigningKey)
	if err != nil {
		a.log.Error("Failed to sign session", "error", err)
		return models.ErrInternalServer
	}

	signatureBytes2, err := base64.StdEncoding.DecodeString(session.Signature)
	if err != nil {
		return models.ErrInvalidCredentials
	}

	if subtle.ConstantTimeCompare(signatureBytes, signatureBytes2) != 1 {
		return models.ErrInvalidCredentials
	}

	return nil
}

func (a *Auth) CreateUser(ctx context.Context, username, password string) (*models.User, error) {
	// Create saltBytes
	saltBytes := make([]byte, saltLenght)
	_, err := rand.Read(saltBytes)
	if err != nil {
		a.log.Error("Failed to generate salt", "error", err)
		return nil, models.ErrInternalServer
	}

	salt := base64.StdEncoding.EncodeToString(saltBytes)

	// Hash password
	hashedPasswordBytes := hashPassword([]byte(password), saltBytes)
	hashedPassword := base64.StdEncoding.EncodeToString(hashedPasswordBytes)

	// Insert user
	user, err := a.repository.Insert(ctx, username, hashedPassword, salt)
	if err != nil {
		a.log.Error("Failed to insert user", "error", err)
		return nil, models.ErrInternalServer
	}

	return &user.User, nil
}
