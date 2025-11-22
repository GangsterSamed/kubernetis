package validators_test

import (
	"github.com/polzovatel/todo-learning/internal/domain/validators"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "valid user",
			email:    "user@example.com",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "invalid email - no @",
			email:    "invalid-email",
			password: "Password123!",
			wantErr:  true,
		},
		{
			name:     "invalid email - no domain",
			email:    "test@",
			password: "Password123!",
			wantErr:  true,
		},
		{
			name:     "password too short",
			email:    "test@example.com",
			password: "Pa123!",
			wantErr:  true,
		},
		{
			name:     "empty email",
			email:    "",
			password: "Password123!",
			wantErr:  true,
		},
		{
			name:     "empty password",
			email:    "test@example.com",
			password: "",
			wantErr:  true,
		},
		{
			name:     "invalid password - no Upper",
			email:    "test@example.com",
			password: "password123!",
			wantErr:  true,
		},
		{
			name:     "invalid password - no !",
			email:    "test@example.com",
			password: "Password123",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateUser(tt.email, tt.password)
			if tt.wantErr {
				assert.Error(t, err, "ValidateUser() should return error for %s", tt.name)
			} else {
				assert.NoError(t, err, "ValidateUser() should not return error for %s", tt.name)
			}
		})
	}
}
