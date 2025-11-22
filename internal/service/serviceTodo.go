package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
	"github.com/polzovatel/todo-learning/internal/models"
	"github.com/polzovatel/todo-learning/internal/repository"
	"github.com/redis/go-redis/v9"
)

type TodoService interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (entities.Todo, error)
	GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*entities.Todo, error)
	GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Todo, error)
	UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, req models.UpdateTodoRequest) (*entities.Todo, error)
	DeleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) error
}

type todoService struct {
	userRepo repository.Store
	todoRepo repository.TodoStore
	cache    *redis.Client
	logger   *slog.Logger
}

func NewTodoService(userRepo repository.Store, todoRepo repository.TodoStore, redis *redis.Client, logger *slog.Logger) TodoService {
	return &todoService{
		userRepo: userRepo,
		todoRepo: todoRepo,
		cache:    redis,
		logger:   logger,
	}
}

func (s *todoService) CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (entities.Todo, error) {
	if _, err := s.userRepo.GetUserById(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: create todo user not found", slog.String("user_id", userID.String()))
			return entities.Todo{}, domain.ErrUserNotFound
		}
		s.logger.Error("service: create todo user lookup failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return entities.Todo{}, err
	}

	todo, err := s.todoRepo.CreateTodo(ctx, userID, title, description)
	if err != nil {
		s.logger.Error("service: create todo failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return entities.Todo{}, err
	}

	if s.cache != nil {
		key := "todo:" + todo.ID.String()
		jsonTodo, err := json.Marshal(todo)
		if err == nil {
			if err := s.cache.Set(ctx, key, jsonTodo, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save cache failed", slog.String("user_id", todo.ID.String()))
			}
		} else {
			s.logger.Error("service: marshal user failed", slog.String("user_id", todo.ID.String()))
		}
	}
	if s.cache != nil {
		listKey := "todos:user:" + userID.String()
		if err := s.cache.Del(ctx, listKey).Err(); err != nil {
			s.logger.Warn("service: failed to invalidate todos list cache", slog.String("user_id", userID.String()))
		}
	}

	s.logger.Info("service: todo created", slog.String("todo_id", todo.ID.String()), slog.String("user_id", userID.String()))
	return todo, nil
}

func (s *todoService) GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*entities.Todo, error) {
	key := "todo:" + todoID.String()
	if s.cache != nil {
		cacheTodo, err := s.cache.Get(ctx, key).Result()
		if err == nil {
			var todo entities.Todo
			s.logger.Info("todo found in cache", slog.String("todo_id", todoID.String()))
			if err := json.Unmarshal([]byte(cacheTodo), &todo); err != nil {
				s.logger.Error("service: unmarshal todo failed", slog.String("todo_id", todoID.String()))
			} else {
				if todo.UserID != userID {
					return nil, domain.ErrForbidden
				}
				return &todo, nil
			}
		}
	}

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

	todoJSON, err := json.Marshal(todo)
	if err == nil {
		if s.cache != nil {
			if err := s.cache.Set(ctx, key, todoJSON, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save todo failed", slog.String("todo_id", todo.ID.String()))
			}
		}
		s.logger.Info("service: save todo", slog.String("todo_id", todo.ID.String()))
	} else {
		s.logger.Error("service: marshal todo failed", slog.String("todo_id", todo.ID.String()))
	}

	s.logger.Info("service: todo found", slog.String("todo_id", todo.ID.String()))
	return todo, nil
}

func (s *todoService) GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Todo, error) {
	if s.cache != nil {
		key := "todos:user:" + userID.String()
		cacheTodo, err := s.cache.Get(ctx, key).Result()
		if err == nil {
			var todos []entities.Todo
			s.logger.Info("todo found in cache", slog.String("user_id", userID.String()))
			if err = json.Unmarshal([]byte(cacheTodo), &todos); err != nil {
				s.logger.Error("service: unmarshal todo failed", slog.String("user_id", userID.String()))
			} else {
				return todos, nil
			}
		}
	}

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

	// Сохраняем в кэш
	if s.cache != nil {
		key := "todos:user:" + userID.String()
		todosJSON, err := json.Marshal(todos)
		if err == nil {
			if err := s.cache.Set(ctx, key, todosJSON, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save todos cache failed", slog.String("user_id", userID.String()))
			} else {
				s.logger.Info("service: todos cached", slog.String("user_id", userID.String()))
			}
		} else {
			s.logger.Error("service: marshal todos failed", slog.String("user_id", userID.String()))
		}
	}

	return todos, nil
}

func (s *todoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, req models.UpdateTodoRequest) (*entities.Todo, error) {
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

	if s.cache != nil {
		key := "todo:" + todo.ID.String()
		if err := s.cache.Del(ctx, key).Err(); err != nil {
			s.logger.Error("service: delete cache failed", slog.String("todo_id", todo.ID.String()))
		}
		jsonTodo, err := json.Marshal(todo)
		if err == nil {
			if err := s.cache.Set(ctx, key, jsonTodo, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save cache failed", slog.String("todo_id", todo.ID.String()))
			}
		} else {
			s.logger.Error("service: marshal todo failed", slog.String("todo_id", todo.ID.String()))
		}
	}
	if s.cache != nil {
		listKey := "todos:user:" + userID.String()
		if err := s.cache.Del(ctx, listKey).Err(); err != nil {
			s.logger.Warn("service: failed to invalidate todos list cache", slog.String("user_id", userID.String()))
		}
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

	if s.cache != nil {
		key := "todo:" + todoID.String()
		if err := s.cache.Del(ctx, key).Err(); err != nil {
			s.logger.Error("service: delete cache failed", slog.String("todo_id", todoID.String()))
		}
	}
	if s.cache != nil {
		listKey := "todos:user:" + userID.String()
		if err := s.cache.Del(ctx, listKey).Err(); err != nil {
			s.logger.Warn("service: failed to invalidate todos list cache", slog.String("user_id", userID.String()))
		}
	}

	s.logger.Info("service: todo deleted", slog.String("todo_id", todoID.String()))
	return nil
}
