package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
	"github.com/polzovatel/todo-learning/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*entities.User, error)
	GetAllUsers(ctx context.Context) ([]entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error
}

type UserService struct {
	store  repository.Store
	cache  *redis.Client
	logger *slog.Logger
}

func NewService(store repository.Store, redis *redis.Client, logger *slog.Logger) Service {
	return &UserService{
		store:  store,
		cache:  redis,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error) {
	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		s.logger.Warn("service: email already taken", slog.String("email", email))
		return entities.User{}, domain.ErrEmailTaken
	}

	user, err := s.store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		s.logger.Error("service: create user failed", slog.String("email", email), slog.Any("error", err))
		return entities.User{}, err
	}

	if s.cache != nil {
		key := "user:" + user.ID.String()
		userJSON, err := json.Marshal(user)
		if err == nil {
			if err := s.cache.Set(ctx, key, userJSON, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save cache failed", slog.String("user_id", user.ID.String()))
			}
		} else {
			s.logger.Error("service: marshal user failed", slog.String("user_id", user.ID.String()))
		}
	}

	s.logger.Info("service: user created", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found by email", slog.String("email", email))
		} else {
			s.logger.Error("service: get user by email failed", slog.String("email", email), slog.Any("error", err))
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (*entities.User, error) {
	// 1. Пробуем получить из кэша
	key := "user:" + userId.String()
	if s.cache != nil {
		cachedUser, err := s.cache.Get(ctx, key).Result()
		if err == nil {
			var user entities.User
			s.logger.Info("user found in cache", slog.String("user_id", userId.String()))
			err = json.Unmarshal([]byte(cachedUser), &user)
			if err != nil {
				s.logger.Error("service: unmarshal user failed", slog.String("user_id", userId.String()))
			} else {
				return &user, nil
			}
		}
	}

	// 2. Не найдено в кэше - идём в БД
	user, err := s.store.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found by id", slog.String("user_id", userId.String()))
		} else {
			s.logger.Error("service: get user by id failed", slog.String("user_id", userId.String()), slog.Any("error", err))
		}
		return nil, err
	}

	// 3. Сохраняем в кэш на 5 минут
	userJSON, err := json.Marshal(user)
	if err == nil {
		if s.cache != nil {
			if err := s.cache.Set(ctx, key, userJSON, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save user failed", slog.String("user_id", userId.String()))
			}
			s.logger.Info("service: save user", slog.String("user_id", userId.String()))
		}
	} else {
		s.logger.Error("service: marshal user failed", slog.String("user_id", userId.String()))
	}

	s.logger.Info("service: user found", slog.String("user_id", userId.String()))
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]entities.User, error) {
	users, err := s.store.GetAllUsers(ctx)
	if err != nil {
		s.logger.Error("service: list users failed", slog.Any("error", err))
		return nil, err
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found for update", slog.String("user_id", user.ID.String()))
		} else {
			s.logger.Error("service: update user failed", slog.String("user_id", user.ID.String()), slog.Any("error", err))
		}
		return nil, err
	}

	if s.cache != nil {
		key := "user:" + user.ID.String()
		if err := s.cache.Del(ctx, key).Err(); err != nil {
			s.logger.Error("service: delete cache failed", slog.String("user_id", user.ID.String()))
		}
		jsonUser, err := json.Marshal(updated)
		if err == nil {
			if err := s.cache.Set(ctx, key, jsonUser, 5*time.Minute).Err(); err != nil {
				s.logger.Error("service: save cache failed", slog.String("user_id", user.ID.String()))
			}
		} else {
			s.logger.Error("service: marshal user failed", slog.String("user_id", user.ID.String()))
		}
	}

	s.logger.Info("service: user updated", slog.String("user_id", updated.ID.String()))
	return updated, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	if err := s.store.DeleteUser(ctx, userId); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found for delete", slog.String("user_id", userId.String()))
		} else {
			s.logger.Error("service: delete user failed", slog.String("user_id", userId.String()), slog.Any("error", err))
		}
		return err
	}

	if s.cache != nil {
		key := "user:" + userId.String()
		if err := s.cache.Del(ctx, key).Err(); err != nil {
			s.logger.Error("service: delete cache failed", slog.String("user_id", userId.String()))
		}
	}

	s.logger.Info("service: user deleted", slog.String("user_id", userId.String()))
	return nil
}
