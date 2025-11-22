package validators

import (
	"errors"
	"net/mail"
	"unicode"
)

var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidPassword  = errors.New("password must be at least 8 characters")
	ErrPasswordNoLower  = errors.New("password must have at least one lowercase letter")
	ErrPasswordNoUpper  = errors.New("password must have at least one uppercase letter")
	ErrPasswordNoSymbol = errors.New("password must have at least one symbol '!'")
	ErrPasswordNoNumber = errors.New("password must have at least one number")
)

func ValidateUser(email, password string) error {
	if !isEmailValid(email) {
		return ErrInvalidEmail
	}
	return ValidatePassword(password)
}

func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	var (
		isLower   bool
		isUpper   bool
		isDigit   bool
		isSpecial bool
	)
	for _, n := range password {
		switch {
		case unicode.IsLower(n):
			isLower = true
		case unicode.IsUpper(n):
			isUpper = true
		case unicode.IsDigit(n):
			isDigit = true
		case n == '!':
			isSpecial = true
		}
	}
	if !isLower {
		return ErrPasswordNoLower
	}
	if !isUpper {
		return ErrPasswordNoUpper
	}
	if !isDigit {
		return ErrPasswordNoNumber
	}
	if !isSpecial {
		return ErrPasswordNoSymbol
	}
	return nil
}
