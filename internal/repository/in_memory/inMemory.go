package in_memory

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
)

type InMemoryRepository struct {
	users     map[uuid.UUID]*models.User
	emailToID map[string]uuid.UUID
	todos     map[uuid.UUID]*models.Todo
	logger    *slog.Logger
}

func NewInMemoryRepository(logger *slog.Logger) *InMemoryRepository {
	return &InMemoryRepository{
		users:     make(map[uuid.UUID]*models.User),
		emailToID: make(map[string]uuid.UUID),
		todos:     make(map[uuid.UUID]*models.Todo),
		logger:    logger,
	}
}

func (r *InMemoryRepository) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	r.users[user.ID] = user
	r.emailToID[user.Email] = user.ID

	if r.logger != nil {
		r.logger.Info("memory: user created", slog.String("user_id", user.ID.String()))
	}
	return *user, nil
}

func (r *InMemoryRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	userID, ok := r.emailToID[email]
	if !ok {
		if r.logger != nil {
			r.logger.Warn("memory: user not found by email", slog.String("email", email))
		}
		return nil, domain.ErrUserNotFound
	}

	return r.GetUserById(ctx, userID)
}

func (r *InMemoryRepository) GetUserById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, ok := r.users[userID]
	if !ok {
		if r.logger != nil {
			r.logger.Warn("memory: user not found by id", slog.String("user_id", userID.String()))
		}
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *InMemoryRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users := make([]models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}

	return users, nil
}

func (r *InMemoryRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	oldUser, err := r.GetUserById(ctx, user.ID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if oldUser.Email != user.Email {
		delete(r.emailToID, oldUser.Email)
	}

	r.users[user.ID] = user
	r.emailToID[user.Email] = user.ID

	if r.logger != nil {
		r.logger.Info("memory: user updated", slog.String("user_id", user.ID.String()))
	}
	return user, nil
}

func (r *InMemoryRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := r.GetUserById(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	delete(r.users, userID)
	delete(r.emailToID, user.Email)

	if r.logger != nil {
		r.logger.Info("memory: user deleted", slog.String("user_id", userID.String()))
	}
	return nil
}
