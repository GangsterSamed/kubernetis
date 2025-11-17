package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
	"github.com/polzovatel/todo-learning/internal/repository"
)

type Service interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error
}

type UserService struct {
	store  repository.Store
	logger *slog.Logger
}

func NewService(store repository.Store, logger *slog.Logger) Service {
	return &UserService{
		store:  store,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		s.logger.Warn("service: email already taken", slog.String("email", email))
		return models.User{}, domain.ErrEmailTaken
	}

	user, err := s.store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		s.logger.Error("service: create user failed", slog.String("email", email), slog.Any("error", err))
		return models.User{}, err
	}

	s.logger.Info("service: user created", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
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

func (s *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user, err := s.store.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found by id", slog.String("user_id", userId.String()))
		} else {
			s.logger.Error("service: get user by id failed", slog.String("user_id", userId.String()), slog.Any("error", err))
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.store.GetAllUsers(ctx)
	if err != nil {
		s.logger.Error("service: list users failed", slog.Any("error", err))
		return nil, err
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("service: user not found for update", slog.String("user_id", user.ID.String()))
		} else {
			s.logger.Error("service: update user failed", slog.String("user_id", user.ID.String()), slog.Any("error", err))
		}
		return nil, err
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
	s.logger.Info("service: user deleted", slog.String("user_id", userId.String()))
	return nil
}
