package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	app2 "github.com/polzovatel/todo-learning/cmd/app"
	"github.com/polzovatel/todo-learning/config"
	"github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/internal/database"
	"github.com/polzovatel/todo-learning/internal/repository"
	"github.com/polzovatel/todo-learning/internal/repository/in_memory"
	"github.com/polzovatel/todo-learning/internal/repository/postgres"
	"github.com/polzovatel/todo-learning/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadCFG()
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})).
			Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	appLogger := logger.SetupLogger(cfg.LogLevel, cfg.LogFormat)

	singer, err := auth.NewJWTSigner(cfg)
	if err != nil {
		appLogger.Error("failed to create JWT signer", slog.Any("error", err))
		os.Exit(1)
	}

	redisClient := database.NewRedisClient(cfg, appLogger)
	if redisClient == nil {
		appLogger.Warn("failed to create redis client, continuing without cache")
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	userRepo, todoRepo, cleanup, err := initRepository(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Error("failed to init repository", slog.Any("error", err))
		os.Exit(1)
	}

	app := app2.NewApp(appLogger, userRepo, todoRepo, redisClient, singer)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: app.Router,
	}

	if err := app.Run(server, cleanup); err != nil {
		appLogger.Error("HTTP server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func initRepository(ctx context.Context, cfg *config.Config, logger *slog.Logger) (repository.Store, repository.TodoStore, func(), error) {
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Warn("database connection failed, falling back to in-memory", slog.Any("error", err))
		repo := in_memory.NewInMemoryRepository(logger)
		return repo, repo, func() {}, nil
	}

	if err := database.RunMigrations(ctx, pool); err != nil {
		logger.Error("migrations failed, falling back to in-memory", slog.Any("error", err))
		repo := in_memory.NewInMemoryRepository(logger)
		pool.Close()
		return repo, repo, func() {}, nil
	}

	repo := postgres.NewPostgresRepository(pool, logger)
	cleanup := func() { pool.Close() }
	return repo, repo, cleanup, nil
}
