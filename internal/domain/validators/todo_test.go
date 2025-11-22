package validators_test

import (
	"testing"

	"github.com/polzovatel/todo-learning/internal/domain/validators"
	"github.com/stretchr/testify/assert"
)

func TestValidateTodo(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{
			name:    "valid title",
			title:   "Купить молоко",
			wantErr: false,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: true,
		},
		{
			name:    "title with only spaces",
			title:   "   ",
			wantErr: true,
		},
		{
			name:    "title with newline",
			title:   "\n\t",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateTodo(tt.title)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
