package domain

import "errors"

// User errors
var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

// Todo errors
var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrForbidden    = errors.New("forbidden")
)
