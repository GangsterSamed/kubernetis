package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
	"github.com/polzovatel/todo-learning/internal/repository"
)

type TodoService interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (models.Todo, error)
	GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*models.Todo, error)
	GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]models.Todo, error)
	UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, req models.UpdateTodoRequest) (*models.Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) error
}

type todoService struct {
	userRepo repository.Store
	todoRepo repository.TodoStore
	logger   *slog.Logger
}

func NewTodoService(userRepo repository.Store, todoRepo repository.TodoStore, logger *slog.Logger) TodoService {
	return &todoService{
		userRepo: userRepo,
		todoRepo: todoRepo,
		logger:   logger,
	}
}

func (s *todoService) CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (models.Todo, error) {
	if _, err := s.userRepo.GetUserById(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: create todo user not found", slog.String("user_id", userID.String()))
			return models.Todo{}, domain.ErrUserNotFound
		}
		s.logger.Error("service: create todo user lookup failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return models.Todo{}, err
	}

	todo, err := s.todoRepo.CreateTodo(ctx, userID, title, description)
	if err != nil {
		s.logger.Error("service: create todo failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return models.Todo{}, err
	}

	s.logger.Info("service: todo created", slog.String("todo_id", todo.ID.String()), slog.String("user_id", userID.String()))
	return todo, nil
}

func (s *todoService) GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*models.Todo, error) {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			s.logger.Warn("service: todo not found", slog.String("todo_id", todoID.String()))
			return nil, domain.ErrTodoNotFound
		}
		s.logger.Error("service: get todo failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return nil, err
	}

	if todo.UserID != userID {
		s.logger.Warn("service: todo access forbidden", slog.String("todo_id", todoID.String()), slog.String("user_id", userID.String()))
		return nil, domain.ErrForbidden
	}

	return todo, nil
}

func (s *todoService) GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	if _, err := s.userRepo.GetUserById(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: list todos user not found", slog.String("user_id", userID.String()))
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("service: list todos user lookup failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return nil, err
	}

	todos, err := s.todoRepo.GetTodoByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("service: list todos failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return nil, err
	}
	return todos, nil
}

func (s *todoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, req models.UpdateTodoRequest) (*models.Todo, error) {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			s.logger.Warn("service: todo not found for update", slog.String("todo_id", todoID.String()))
			return nil, domain.ErrTodoNotFound
		}
		s.logger.Error("service: get todo before update failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return nil, err
	}

	if todo.UserID != userID {
		s.logger.Warn("service: todo update forbidden", slog.String("todo_id", todoID.String()), slog.String("user_id", userID.String()))
		return nil, domain.ErrForbidden
	}

	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	todo.UpdatedAt = time.Now()

	if _, err := s.todoRepo.UpdateTodo(ctx, todo); err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			s.logger.Warn("service: todo not found during update write", slog.String("todo_id", todoID.String()))
			return nil, domain.ErrTodoNotFound
		}
		s.logger.Error("service: update todo failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return nil, err
	}

	s.logger.Info("service: todo updated", slog.String("todo_id", todo.ID.String()))
	return todo, nil
}

func (s *todoService) DeleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) error {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			s.logger.Warn("service: todo not found for delete", slog.String("todo_id", todoID.String()))
			return domain.ErrTodoNotFound
		}
		s.logger.Error("service: get todo before delete failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return err
	}

	if todo.UserID != userID {
		s.logger.Warn("service: todo delete forbidden", slog.String("todo_id", todoID.String()), slog.String("user_id", userID.String()))
		return domain.ErrForbidden
	}

	if err := s.todoRepo.DeleteTodo(ctx, todoID); err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			s.logger.Warn("service: todo not found during delete write", slog.String("todo_id", todoID.String()))
			return domain.ErrTodoNotFound
		}
		s.logger.Error("service: delete todo failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return err
	}

	s.logger.Info("service: todo deleted", slog.String("todo_id", todoID.String()))
	return nil
}
