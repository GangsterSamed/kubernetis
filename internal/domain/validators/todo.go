package validators

import (
	"errors"
	"strings"
)

var (
	ErrTitleEmpty = errors.New("Title cannot be empty")
)

func ValidateTodo(title string) error {
	if len(strings.TrimSpace(title)) == 0 {
		return ErrTitleEmpty
	}
	return nil
}
