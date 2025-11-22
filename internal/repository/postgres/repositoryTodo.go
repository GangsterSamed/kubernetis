package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/domain/entities"
)

func (r *PostgresRepository) CreateTodo(ctx context.Context, userID uuid.UUID, title, description string) (entities.Todo, error) {
	todoID := uuid.New()
	const q = `INSERT INTO todos (id, user_id, title, description) VALUES ($1, $2, $3, $4) RETURNING id, user_id, title, description, completed, created_at, updated_at`

	var todo entities.Todo
	if err := r.pool.QueryRow(ctx, q, todoID, userID, title, description).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
		r.logger.Error("postgres: create todo failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return entities.Todo{}, err
	}

	r.logger.Info("postgres: todo created", slog.String("todo_id", todo.ID.String()), slog.String("user_id", userID.String()))
	return todo, nil
}

func (r *PostgresRepository) GetTodoByID(ctx context.Context, todoID uuid.UUID) (*entities.Todo, error) {
	const q = `SELECT id, user_id, title, description, completed, created_at, updated_at FROM todos WHERE id = $1`

	var todo entities.Todo
	if err := r.pool.QueryRow(ctx, q, todoID).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("postgres: todo not found", slog.String("todo_id", todoID.String()))
			return nil, domain.ErrTodoNotFound
		}
		r.logger.Error("postgres: get todo failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return nil, err
	}

	return &todo, nil
}

func (r *PostgresRepository) GetTodoByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Todo, error) {
	const q = `SELECT id, user_id, title, description, completed, created_at, updated_at FROM todos WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		r.logger.Error("postgres: list todos failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	todos := make([]entities.Todo, 0)
	for rows.Next() {
		var todo entities.Todo
		if err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			r.logger.Error("postgres: scan todo failed", slog.Any("error", err))
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("postgres: rows iteration failed", slog.Any("error", err))
		return nil, err
	}

	return todos, nil
}

func (r *PostgresRepository) UpdateTodo(ctx context.Context, todo *entities.Todo) (*entities.Todo, error) {
	const q = `UPDATE todos SET user_id = $1, title = $2, description = $3, completed = $4, updated_at = NOW() WHERE id = $5 RETURNING id, user_id, title, description, completed, created_at, updated_at`

	if err := r.pool.QueryRow(ctx, q, todo.UserID, todo.Title, todo.Description, todo.Completed, todo.ID).
		Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("postgres: update todo target not found", slog.String("todo_id", todo.ID.String()))
			return nil, domain.ErrTodoNotFound
		}
		r.logger.Error("postgres: update todo failed", slog.String("todo_id", todo.ID.String()), slog.Any("error", err))
		return nil, err
	}

	r.logger.Info("postgres: todo updated", slog.String("todo_id", todo.ID.String()))
	return todo, nil
}

func (r *PostgresRepository) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	const q = `DELETE FROM todos WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, q, todoID)
	if err != nil {
		r.logger.Error("postgres: delete todo failed", slog.String("todo_id", todoID.String()), slog.Any("error", err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		r.logger.Warn("postgres: delete todo target not found", slog.String("todo_id", todoID.String()))
		return domain.ErrTodoNotFound
	}

	r.logger.Info("postgres: todo deleted", slog.String("todo_id", todoID.String()))
	return nil
}
