package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
	"github.com/polzovatel/todo-learning/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	t.Run("create user successfully", func(t *testing.T) {
		user, err := service.CreateUser(ctx, "new@example.com", "hashed_password")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "new@example.com", user.Email)
	})

	t.Run("create user with duplicate email", func(t *testing.T) {
		user1, err1 := service.CreateUser(ctx, "duplicate@example.com", "hashed_password_1")
		require.NoError(t, err1)
		require.NotNil(t, user1)

		_, err2 := service.CreateUser(ctx, "duplicate@example.com", "hashed_password_2")
		assert.Error(t, err2)
		assert.Equal(t, domain.ErrEmailTaken, err2)
	})
}

func TestUserService_GetUserById(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	// Создаем пользователя
	created, err := service.CreateUser(ctx, "new@example.com", "hashed_password")
	require.NoError(t, err)
	require.NotNil(t, created)

	t.Run("get user successfully", func(t *testing.T) {
		user, err := service.GetUserById(ctx, created.ID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, "new@example.com", user.Email)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := service.GetUserById(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	created, err := service.CreateUser(ctx, "new@example.com", "hashed_password")
	require.NoError(t, err)
	require.NotNil(t, created)

	t.Run("get user successfully", func(t *testing.T) {
		user, err := service.GetUserByEmail(ctx, created.Email)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Email, user.Email)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := service.GetUserByEmail(ctx, "notfound@example.com")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	t.Run("get all users successfully", func(t *testing.T) {
		user1, _ := service.CreateUser(ctx, "user1@example.com", "hash1")
		user2, _ := service.CreateUser(ctx, "user2@example.com", "hash2")

		users, err := service.GetAllUsers(ctx)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)

		userIDs := make(map[uuid.UUID]bool)
		for _, user := range users {
			userIDs[user.ID] = true
		}
		assert.True(t, userIDs[user1.ID])
		assert.True(t, userIDs[user2.ID])
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	created, err := service.CreateUser(ctx, "update@example.com", "hash")
	require.NoError(t, err)

	t.Run("update user successfully", func(t *testing.T) {
		created.Email = "updated@example.com"
		created.PasswordHash = "new_hash"

		updated, err := service.UpdateUser(ctx, &created)

		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "updated@example.com", updated.Email)
		assert.Equal(t, "new_hash", updated.PasswordHash)
	})

	t.Run("update non-existing user", func(t *testing.T) {
		nonExistent := &entities.User{
			ID:           uuid.New(),
			Email:        "nonexistent@example.com",
			PasswordHash: "hash",
		}

		_, err := service.UpdateUser(ctx, nonExistent)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	mockStore := mocks.NewMockStore()
	service := NewService(mockStore, nil, slog.Default())

	t.Run("delete user successfully", func(t *testing.T) {
		created, err := service.CreateUser(ctx, "delete@example.com", "hash")
		require.NoError(t, err)

		err = service.DeleteUser(ctx, created.ID)
		require.NoError(t, err)

		// Проверяем что пользователь удален
		_, err = service.GetUserById(ctx, created.ID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := service.DeleteUser(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}
