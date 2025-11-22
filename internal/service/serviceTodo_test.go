package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
	"github.com/polzovatel/todo-learning/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoService_CreateTodo(t *testing.T) {
	ctx := context.Background()
	mockUserStore := mocks.NewMockStore()
	mockTodoStore := mocks.NewMockTodoStore()
	service := NewTodoService(mockUserStore, mockTodoStore, nil, slog.Default())

	// Создаем пользователя
	user, err := mockUserStore.CreateUser(ctx, "todo@example.com", "hash")
	require.NoError(t, err)

	t.Run("create todo successfully", func(t *testing.T) {
		todo, err := service.CreateTodo(ctx, user.ID, "Test Todo", "Description")

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, todo.ID)
		assert.Equal(t, user.ID, todo.UserID)
		assert.Equal(t, "Test Todo", todo.Title)
		assert.Equal(t, "Description", todo.Description)
		assert.False(t, todo.Completed)
	})

	t.Run("create todo with non-existing user", func(t *testing.T) {
		_, err := service.CreateTodo(ctx, uuid.New(), "Test Todo", "Description")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestTodoService_GetTodoByID(t *testing.T) {
	ctx := context.Background()
	mockUserStore := mocks.NewMockStore()
	mockTodoStore := mocks.NewMockTodoStore()
	service := NewTodoService(mockUserStore, mockTodoStore, nil, slog.Default())

	user, _ := mockUserStore.CreateUser(ctx, "get@example.com", "hash")
	created, _ := service.CreateTodo(ctx, user.ID, "Test Todo", "Description")

	t.Run("get todo by id successfully", func(t *testing.T) {
		todo, err := service.GetTodoByID(ctx, created.ID, user.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, todo.ID)
		assert.Equal(t, "Test Todo", todo.Title)
		assert.Equal(t, user.ID, todo.UserID)
	})

	t.Run("get todo with wrong user", func(t *testing.T) {
		otherUser, _ := mockUserStore.CreateUser(ctx, "other@example.com", "hash")

		_, err := service.GetTodoByID(ctx, created.ID, otherUser.ID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrForbidden, err)
	})

	t.Run("get non-existing todo", func(t *testing.T) {
		_, err := service.GetTodoByID(ctx, uuid.New(), user.ID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}

func TestTodoService_GetTodoByUserID(t *testing.T) {
	ctx := context.Background()
	mockUserStore := mocks.NewMockStore()
	mockTodoStore := mocks.NewMockTodoStore()
	service := NewTodoService(mockUserStore, mockTodoStore, nil, slog.Default())

	user1, _ := mockUserStore.CreateUser(ctx, "user1@example.com", "hash")
	user2, _ := mockUserStore.CreateUser(ctx, "user2@example.com", "hash")

	todo1, _ := service.CreateTodo(ctx, user1.ID, "Todo 1", "")
	todo2, _ := service.CreateTodo(ctx, user1.ID, "Todo 2", "")
	_, _ = service.CreateTodo(ctx, user2.ID, "Todo 3", "")

	t.Run("get todos for user successfully", func(t *testing.T) {
		todos, err := service.GetTodoByUserID(ctx, user1.ID)

		require.NoError(t, err)
		assert.Len(t, todos, 2)

		todoIDs := make(map[uuid.UUID]bool)
		for _, todo := range todos {
			todoIDs[todo.ID] = true
			assert.Equal(t, user1.ID, todo.UserID)
		}
		assert.True(t, todoIDs[todo1.ID])
		assert.True(t, todoIDs[todo2.ID])
	})

	t.Run("get todos for non-existing user", func(t *testing.T) {
		_, err := service.GetTodoByUserID(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestTodoService_UpdateTodo(t *testing.T) {
	ctx := context.Background()
	mockUserStore := mocks.NewMockStore()
	mockTodoStore := mocks.NewMockTodoStore()
	service := NewTodoService(mockUserStore, mockTodoStore, nil, slog.Default())

	user, _ := mockUserStore.CreateUser(ctx, "update@example.com", "hash")
	created, _ := service.CreateTodo(ctx, user.ID, "Original Title", "Original Description")

	t.Run("update todo successfully", func(t *testing.T) {
		newTitle := "Updated Title"
		newDescription := "Updated Description"
		completed := true

		req := models.UpdateTodoRequest{
			Title:       &newTitle,
			Description: &newDescription,
			Completed:   &completed,
		}

		updated, err := service.UpdateTodo(ctx, created.ID, user.ID, req)

		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated Description", updated.Description)
		assert.True(t, updated.Completed)
	})

	t.Run("update todo with wrong user", func(t *testing.T) {
		otherUser, _ := mockUserStore.CreateUser(ctx, "other@example.com", "hash")
		newTitle := "Hacked Title"

		req := models.UpdateTodoRequest{
			Title: &newTitle,
		}

		_, err := service.UpdateTodo(ctx, created.ID, otherUser.ID, req)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrForbidden, err)
	})

	t.Run("update non-existing todo", func(t *testing.T) {
		newTitle := "New Title"
		req := models.UpdateTodoRequest{
			Title: &newTitle,
		}

		_, err := service.UpdateTodo(ctx, uuid.New(), user.ID, req)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}

func TestTodoService_DeleteTodo(t *testing.T) {
	ctx := context.Background()
	mockUserStore := mocks.NewMockStore()
	mockTodoStore := mocks.NewMockTodoStore()
	service := NewTodoService(mockUserStore, mockTodoStore, nil, slog.Default())

	t.Run("delete todo successfully", func(t *testing.T) {
		user, _ := mockUserStore.CreateUser(ctx, "delete@example.com", "hash")
		created, _ := service.CreateTodo(ctx, user.ID, "To Delete", "")

		err := service.DeleteTodo(ctx, created.ID, user.ID)
		require.NoError(t, err)

		// Проверяем что todo удален
		_, err = service.GetTodoByID(ctx, created.ID, user.ID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})

	t.Run("delete todo with wrong user", func(t *testing.T) {
		user1, _ := mockUserStore.CreateUser(ctx, "user1@example.com", "hash")
		user2, _ := mockUserStore.CreateUser(ctx, "user2@example.com", "hash")
		created, _ := service.CreateTodo(ctx, user1.ID, "To Delete", "")

		err := service.DeleteTodo(ctx, created.ID, user2.ID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrForbidden, err)
	})

	t.Run("delete non-existing todo", func(t *testing.T) {
		user, _ := mockUserStore.CreateUser(ctx, "delete2@example.com", "hash")

		err := service.DeleteTodo(ctx, uuid.New(), user.ID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}
