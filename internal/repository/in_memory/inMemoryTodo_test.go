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

func TestInMemoryTodo_CreateTodo(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	// Создаем пользователя для todo
	user, err := repo.CreateUser(ctx, "todo@example.com", "hash")
	require.NoError(t, err)

	t.Run("create todo successfully", func(t *testing.T) {
		todo, err := repo.CreateTodo(ctx, user.ID, "Test Todo", "Description")

		require.NoError(t, err)
		require.NotEmpty(t, todo)
		assert.NotEqual(t, uuid.Nil, todo.ID)
		assert.Equal(t, user.ID, todo.UserID)
		assert.Equal(t, "Test Todo", todo.Title)
		assert.Equal(t, "Description", todo.Description)
		assert.False(t, todo.Completed)
	})
}

func TestInMemoryRepository_GetTodoByUserID(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	user1, _ := repo.CreateUser(ctx, "user1@example.com", "hash")
	user2, _ := repo.CreateUser(ctx, "user2@example.com", "hash")

	// Создаем todos для user1
	todo1, _ := repo.CreateTodo(ctx, user1.ID, "Todo 1", "")
	todo2, _ := repo.CreateTodo(ctx, user1.ID, "Todo 2", "")

	// Создаем todo для user2
	_, _ = repo.CreateTodo(ctx, user2.ID, "Todo 3", "")

	t.Run("get todos for user1", func(t *testing.T) {
		todos, err := repo.GetTodoByUserID(ctx, user1.ID)

		require.NoError(t, err)
		assert.Len(t, todos, 2)
		// Проверяем что вернулись правильные todos
		todoIDs := make(map[uuid.UUID]bool)
		for _, todo := range todos {
			todoIDs[todo.ID] = true
			assert.Equal(t, user1.ID, todo.UserID)
		}
		assert.True(t, todoIDs[todo1.ID])
		assert.True(t, todoIDs[todo2.ID])
	})

	t.Run("get todos for user with no todos", func(t *testing.T) {
		newUser, _ := repo.CreateUser(ctx, "new@example.com", "hash")
		todos, err := repo.GetTodoByUserID(ctx, newUser.ID)

		require.NoError(t, err)
		assert.Empty(t, todos)
	})
}

func TestInMemoryRepository_GetTodoByID(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "getbyid@example.com", "hash")
	created, _ := repo.CreateTodo(ctx, user.ID, "Test Todo", "Description")

	t.Run("get todo by id successfully", func(t *testing.T) {
		todo, err := repo.GetTodoByID(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, todo.ID)
		assert.Equal(t, "Test Todo", todo.Title)
		assert.Equal(t, user.ID, todo.UserID)
	})

	t.Run("get non-existing todo", func(t *testing.T) {
		_, err := repo.GetTodoByID(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}

func TestInMemoryRepository_UpdateTodo(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "update@example.com", "hash")
	created, _ := repo.CreateTodo(ctx, user.ID, "Original Title", "Original Description")

	t.Run("update todo successfully", func(t *testing.T) {
		created.Title = "Updated Title"
		created.Description = "Updated Description"
		created.Completed = true

		updated, err := repo.UpdateTodo(ctx, &created)

		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated Description", updated.Description)
		assert.True(t, updated.Completed)
	})

	t.Run("update non-existing todo", func(t *testing.T) {
		nonExistent := &entities.Todo{
			ID:     uuid.New(),
			UserID: user.ID,
			Title:  "Non-existent",
		}

		_, err := repo.UpdateTodo(ctx, nonExistent)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}

func TestInMemoryRepository_DeleteTodo(t *testing.T) {
	repo := in_memory.NewInMemoryRepository(slog.Default())
	ctx := context.Background()

	t.Run("delete todo successfully", func(t *testing.T) {
		user, _ := repo.CreateUser(ctx, "delete@example.com", "hash")
		created, _ := repo.CreateTodo(ctx, user.ID, "To Delete", "")

		err := repo.DeleteTodo(ctx, created.ID)
		require.NoError(t, err)

		// Проверяем что todo удален
		_, err = repo.GetTodoByID(ctx, created.ID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})

	t.Run("delete non-existing todo", func(t *testing.T) {
		err := repo.DeleteTodo(ctx, uuid.New())

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTodoNotFound, err)
	})
}
