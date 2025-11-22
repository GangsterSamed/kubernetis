package mappers

import (
	"github.com/polzovatel/todo-learning/internal/domain/entities"
	"github.com/polzovatel/todo-learning/internal/models"
)

func UserToDTO(user entities.User) models.UserResponse {
	return models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
