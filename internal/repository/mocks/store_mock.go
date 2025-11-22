package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
)

type MockStore struct {
	Users        map[uuid.UUID]*entities.User
	UsersByEmail map[string]*entities.User
}

func NewMockStore() *MockStore {
	return &MockStore{
		Users:        make(map[uuid.UUID]*entities.User),
		UsersByEmail: make(map[string]*entities.User),
	}
}

func (s *MockStore) CreateUser(ctx context.Context, email, passwordHash string) (entities.User, error) {
	user := entities.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.Users[user.ID] = &user
	s.UsersByEmail[user.Email] = &user
	return user, nil
}

func (s *MockStore) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, ok := s.UsersByEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *MockStore) GetUserById(ctx context.Context, userId uuid.UUID) (*entities.User, error) {
	user, ok := s.Users[userId]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *MockStore) GetAllUsers(ctx context.Context) ([]entities.User, error) {
	var users []entities.User
	for _, user := range s.Users {
		users = append(users, *user)
	}
	return users, nil
}

func (s *MockStore) UpdateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	if _, ok := s.Users[user.ID]; !ok {
		return nil, domain.ErrUserNotFound
	}
	oldUser := s.Users[user.ID]
	if oldUser.Email != user.Email {
		delete(s.UsersByEmail, oldUser.Email)
		s.UsersByEmail[user.Email] = user
	}
	s.Users[user.ID] = user
	return user, nil
}

func (s *MockStore) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	user, ok := s.Users[userId]
	if !ok {
		return domain.ErrUserNotFound
	}
	delete(s.Users, userId)
	// Также удаляем из UsersByEmail
	delete(s.UsersByEmail, user.Email)
	return nil
}
