package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
)

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserById(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetAllUsers(ctx context.Context) ([]entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type TodoStore interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (entities.Todo, error)
	GetTodoByID(ctx context.Context, todoID uuid.UUID) (*entities.Todo, error)
	GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Todo, error)
	UpdateTodo(ctx context.Context, todo *entities.Todo) (*entities.Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
}
