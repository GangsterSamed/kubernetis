package main

import (
	"context"
	"os"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/polzovatel/todo-learning/cmd/middleware"
	"github.com/polzovatel/todo-learning/config"
	"github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/internal/controller"
	"github.com/polzovatel/todo-learning/internal/database"
	"github.com/polzovatel/todo-learning/internal/repository"
	"github.com/polzovatel/todo-learning/internal/repository/in_memory"
	"github.com/polzovatel/todo-learning/internal/repository/postgres"
	"github.com/polzovatel/todo-learning/internal/service"
	"github.com/polzovatel/todo-learning/logger"
)

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

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

	userRepo, todoRepo, cleanup, err := initRepository(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Error("failed to init repository", slog.Any("error", err))
		os.Exit(1)
	}
	defer cleanup()

	userService := service.NewService(userRepo, appLogger)
	todoService := service.NewTodoService(userRepo, todoRepo, appLogger)
	contr := controller.NewUserController(userService, singer, appLogger)
	todoContr := controller.NewTodoController(todoService, singer, appLogger)

	r.Use(middleware.RequestLoggerMiddleware(appLogger))
	api := r.Group("/api/v1")
	{
		api.POST("/register", contr.RegisterUser)
		api.POST("/login", contr.LoginUser)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(singer, appLogger))
	{
		protected.GET("/me", contr.GetMe)
		protected.POST("/logout", contr.LogoutUser)
		protected.POST("/todos", todoContr.CreateTodo)
		protected.GET("/todos", todoContr.GetTodos)
		protected.GET("/todos/:id", todoContr.GetTodoByID)
		protected.PUT("/todos/:id", todoContr.UpdateTodo)
		protected.DELETE("/todos/:id", todoContr.DeleteTodo)
	}

	appLogger.Info("HTTP server starting", slog.String("addr", cfg.HTTPAddr))
	if err := r.Run(cfg.HTTPAddr); err != nil {
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
