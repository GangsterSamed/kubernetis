package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
)

type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresRepository(pool *pgxpool.Pool, logger *slog.Logger) *PostgresRepository {
	return &PostgresRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at`

	var user models.User
	if err := r.pool.QueryRow(ctx, q, email, passwordHash).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		r.logger.Error("postgres: create user failed", slog.String("email", email), slog.Any("error", err))
		return models.User{}, err
	}

	r.logger.Info("postgres: user created", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`

	var user models.User
	if err := r.pool.QueryRow(ctx, q, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("postgres: user not found by email", slog.String("email", email))
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("postgres: get user by email failed", slog.String("email", email), slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`

	var user models.User
	if err := r.pool.QueryRow(ctx, q, userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("postgres: user not found by id", slog.String("user_id", userID.String()))
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("postgres: get user by id failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		r.logger.Error("postgres: list users failed", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
			r.logger.Error("postgres: scan user failed", slog.Any("error", err))
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("postgres: rows iteration failed", slog.Any("error", err))
		return nil, err
	}

	return users, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	const q = `UPDATE users SET email = $1, password_hash = $2 WHERE id = $3 RETURNING id, email, password_hash, created_at`

	if err := r.pool.QueryRow(ctx, q, user.Email, user.PasswordHash, user.ID).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("postgres: update user target not found", slog.String("user_id", user.ID.String()))
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("postgres: update user failed", slog.String("user_id", user.ID.String()), slog.Any("error", err))
		return nil, err
	}

	r.logger.Info("postgres: user updated", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		r.logger.Error("postgres: delete user failed", slog.String("user_id", userID.String()), slog.Any("error", err))
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		r.logger.Warn("postgres: delete user target not found", slog.String("user_id", userID.String()))
		return domain.ErrUserNotFound
	}

	r.logger.Info("postgres: user deleted", slog.String("user_id", userID.String()))
	return nil
}
