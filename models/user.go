package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type User struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserCredentials struct {
	User
	Password string
	Salt     string
}

type Session struct {
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
	Signature string    `json:"-"`
}

func (sess *Session) ToCookie() (string, error) {
	payloadBytes, err := json.Marshal(sess)
	if err != nil {
		return "", err
	}

	payload := base64.StdEncoding.EncodeToString(payloadBytes)
	return payload + ":" + sess.Signature, nil
}

func SessionFromCookie(cookie string) (*Session, error) {
	parts := strings.Split(cookie, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid cookie")
	}

	payload, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	sess := &Session{}
	err = json.Unmarshal(payload, sess)
	if err != nil {
		return nil, err
	}

	sess.Signature = parts[1]
	return sess, nil
}
