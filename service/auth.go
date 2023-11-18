package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/viddrobnic/sparovec/config"
	"github.com/viddrobnic/sparovec/models"
)

type AuthUserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.UserCredentials, error)
}

type Auth struct {
	repository AuthUserRepository

	conf *config.Config
	log  *slog.Logger
}

func NewAuth(repository AuthUserRepository, conf *config.Config, log *slog.Logger) *Auth {
	return &Auth{
		repository: repository,
		conf:       conf,
		log:        log,
	}
}

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

func doPasswordsMatch(hashedPassword, salt, password string) bool {
	hashedPasswordBytes, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false
	}

	hashedPasswordBytes2 := hashPassword([]byte(password), saltBytes)

	return subtle.ConstantTimeCompare(hashedPasswordBytes, hashedPasswordBytes2) == 1
}

func signSession(sess *models.Session, signingKey string) ([]byte, error) {
	// Marshal session
	sessBytes, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	// Create signature with HMAC
	sum := hmac.New(sha256.New, []byte(signingKey))
	_, err = sum.Write(sessBytes)
	if err != nil {
		return nil, err
	}

	return sum.Sum(nil), nil
}
