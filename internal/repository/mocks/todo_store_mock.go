package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
)

type MockTodoStore struct {
	Todos map[uuid.UUID]*entities.Todo
}

func NewMockTodoStore() *MockTodoStore {
	return &MockTodoStore{
		Todos: make(map[uuid.UUID]*entities.Todo),
	}
}

func (m *MockTodoStore) CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (entities.Todo, error) {
	todo := entities.Todo{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.Todos[todo.ID] = &todo
	return todo, nil
}

func (m *MockTodoStore) GetTodoByID(ctx context.Context, todoID uuid.UUID) (*entities.Todo, error) {
	todo, ok := m.Todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return todo, nil
}

func (m *MockTodoStore) GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Todo, error) {
	var todos []entities.Todo
	for _, todo := range m.Todos {
		if todo.UserID == userID {
			todos = append(todos, *todo)
		}
	}
	return todos, nil
}

func (m *MockTodoStore) UpdateTodo(ctx context.Context, todo *entities.Todo) (*entities.Todo, error) {
	if _, ok := m.Todos[todo.ID]; !ok {
		return nil, domain.ErrTodoNotFound
	}
	todo.UpdatedAt = time.Now()
	m.Todos[todo.ID] = todo
	return todo, nil
}

func (m *MockTodoStore) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	if _, ok := m.Todos[todoID]; !ok {
		return domain.ErrTodoNotFound
	}
	delete(m.Todos, todoID)
	return nil
}
