package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"

	"github.com/viddrobnic/sparovec/models"
	"golang.org/x/crypto/argon2"
)

func hashPassword(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
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
