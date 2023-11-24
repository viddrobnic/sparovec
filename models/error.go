package models

import (
	"errors"
	"fmt"
)

var (
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("forbidden")
)

type ErrInvalidForm struct {
	Message string
}

func (e *ErrInvalidForm) Error() string {
	return fmt.Sprintf("invalid form: %s", e.Message)
}
