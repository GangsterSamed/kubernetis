package domain

import "errors"

var ErrTodoNotFound = errors.New("todo not found")

var (
	ErrUserNotFound = errors.New("user not found")
	ErrForbidden    = errors.New("forbidden")
	ErrEmailTaken   = errors.New("email already taken")
)
