package in_memory_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
	"github.com/polzovatel/todo-learning/internal/repository/in_memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryUser_CreateUser(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	t.Run("create user successfully", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "test@example.com", "hashed_password")

		require.NoError(t, err)
		assert.NotEmpty(t, user)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "hashed_password", user.PasswordHash)
	})
}

func TestInMemoryUser_GetUserByEmail(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	// Создаем пользователя
	created, err := repo.CreateUser(ctx, "test@example.com", "hash")
	require.NoError(t, err)

	t.Run("get user by email successfully", func(t *testing.T) {
		user, err := repo.GetUserByEmail(ctx, "test@example.com")

		require.NoError(t, err)
		assert.NotEmpty(t, user)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := repo.GetUserByEmail(ctx, "notfound@example.com")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestInMemoryUser_GetUserByID(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "get@example.com", "hash")
	require.NoError(t, err)

	t.Run("get user by id successfully", func(t *testing.T) {
		user, err := repo.GetUserById(ctx, created.ID)

		require.NoError(t, err)
		assert.NotEmpty(t, user)
		assert.Equal(t, created.ID, user.ID)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		_, err := repo.GetUserById(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestInMemoryUser_GetAllUsers(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	t.Run("get all users successfully", func(t *testing.T) {
		user1, _ := repo.CreateUser(ctx, "user1@example.com", "hash1")
		user2, _ := repo.CreateUser(ctx, "user2@example.com", "hash2")

		users, err := repo.GetAllUsers(ctx)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)

		userIDs := make(map[uuid.UUID]bool)
		for _, user := range users {
			userIDs[user.ID] = true
		}
		assert.True(t, userIDs[user1.ID])
		assert.True(t, userIDs[user2.ID])
	})

	t.Run("get all users when empty", func(t *testing.T) {
		emptyRepo := in_memory.NewInMemoryRepository(slog.Default())
		users, err := emptyRepo.GetAllUsers(ctx)

		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestInMemoryUser_UpdateUser(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	created, err := repo.CreateUser(ctx, "update@example.com", "hash")
	require.NoError(t, err)

	t.Run("update user successfully", func(t *testing.T) {
		created.Email = "updated@example.com"
		created.PasswordHash = "new_hash"

		updated, err := repo.UpdateUser(ctx, &created)

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

		_, err := repo.UpdateUser(ctx, nonExistent)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestInMemoryUser_DeleteUser(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	t.Run("delete user successfully", func(t *testing.T) {
		created, err := repo.CreateUser(ctx, "delete@example.com", "hash")
		require.NoError(t, err)

		err = repo.DeleteUser(ctx, created.ID)
		require.NoError(t, err)

		// Проверяем что пользователь удален
		_, err = repo.GetUserById(ctx, created.ID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := repo.DeleteUser(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}
