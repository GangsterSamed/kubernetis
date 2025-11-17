package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/models"
)

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type TodoStore interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (models.Todo, error)
	GetTodoByID(ctx context.Context, todoID uuid.UUID) (*models.Todo, error)
	GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]models.Todo, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) (*models.Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
}
