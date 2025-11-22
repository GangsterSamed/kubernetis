package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polzovatel/todo-learning/cmd/middleware"
	"github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/internal/controller"
	"github.com/polzovatel/todo-learning/internal/repository"
	"github.com/polzovatel/todo-learning/internal/service"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Router   *gin.Engine
	logger   *slog.Logger
	signer   *auth.JWTSigner
	userCtrl *controller.UserController
	todoCtrl *controller.TodoController
}

func NewApp(logger *slog.Logger, userRepo repository.Store, todoRepo repository.TodoStore, redisClient *redis.Client, signer *auth.JWTSigner) *App {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLoggerMiddleware(logger))

	userService := service.NewService(userRepo, redisClient, logger)
	todoService := service.NewTodoService(userRepo, todoRepo, redisClient, logger)
	contr := controller.NewUserController(userService, signer, logger)
	todoContr := controller.NewTodoController(todoService, signer, logger)

	app := &App{
		Router:   r,
		logger:   logger,
		signer:   signer,
		userCtrl: contr,
		todoCtrl: todoContr,
	}

	app.SetupRoutes()
	return app
}

func (app *App) SetupRoutes() {
	api := app.Router.Group("/api/v1")
	{
		api.POST("/register", app.userCtrl.RegisterUser)
		api.POST("/login", app.userCtrl.LoginUser)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(app.signer, app.logger))
	{
		protected.GET("/me", app.userCtrl.GetMe)
		protected.POST("/logout", app.userCtrl.LogoutUser)
		protected.POST("/todos", app.todoCtrl.CreateTodo)
		protected.GET("/todos", app.todoCtrl.GetTodos)
		protected.GET("/todos/:id", app.todoCtrl.GetTodoByID)
		protected.PUT("/todos/:id", app.todoCtrl.UpdateTodo)
		protected.DELETE("/todos/:id", app.todoCtrl.DeleteTodo)
	}
}

func (app *App) Run(server *http.Server, cleanup func()) error {
	errChan := make(chan error, 1)
	go func() {
		app.logger.Info("HTTP server starting", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-quit:
		app.logger.Info("shutdown signal received", slog.String("signal", sig.String()))
		return app.gracefulShutdown(server, cleanup)
	}
}

func (app *App) gracefulShutdown(server *http.Server, cleanup func()) error {
	app.logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.logger.Error("server shutdown error", slog.Any("error", err))
		return err
	}

	app.logger.Info("HTTP server stopped")

	if cleanup != nil {
		cleanup()
		app.logger.Info("database connections closed")
	}

	app.logger.Info("graceful shutdown completed")
	return nil
}
