package in_memory

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
)

func (r *InMemoryRepository) CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (models.Todo, error) {
	todo := &models.Todo{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	r.todos[todo.ID] = todo

	if r.logger != nil {
		r.logger.Info("memory: todo created", slog.String("todo_id", todo.ID.String()), slog.String("user_id", userID.String()))
	}
	return *todo, nil
}

func (r *InMemoryRepository) GetTodoByID(ctx context.Context, todoID uuid.UUID) (*models.Todo, error) {
	todo, ok := r.todos[todoID]
	if !ok {
		if r.logger != nil {
			r.logger.Warn("memory: todo not found", slog.String("todo_id", todoID.String()))
		}
		return nil, domain.ErrTodoNotFound
	}

	return todo, nil
}

func (r *InMemoryRepository) GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	todos := make([]models.Todo, 0)
	for _, todo := range r.todos {
		if todo.UserID == userID {
			todos = append(todos, *todo)
		}
	}

	return todos, nil
}

func (r *InMemoryRepository) UpdateTodo(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	if _, ok := r.todos[todo.ID]; !ok {
		if r.logger != nil {
			r.logger.Warn("memory: todo not found for update", slog.String("todo_id", todo.ID.String()))
		}
		return nil, domain.ErrTodoNotFound
	}

	r.todos[todo.ID] = todo

	if r.logger != nil {
		r.logger.Info("memory: todo updated", slog.String("todo_id", todo.ID.String()))
	}
	return todo, nil
}

func (r *InMemoryRepository) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	if _, ok := r.todos[todoID]; !ok {
		if r.logger != nil {
			r.logger.Warn("memory: todo not found for delete", slog.String("todo_id", todoID.String()))
		}
		return domain.ErrTodoNotFound
	}

	delete(r.todos, todoID)

	if r.logger != nil {
		r.logger.Info("memory: todo deleted", slog.String("todo_id", todoID.String()))
	}
	return nil
}
