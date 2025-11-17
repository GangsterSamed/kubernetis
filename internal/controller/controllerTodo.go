package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	auth2 "github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/internal/domain"
	"github.com/polzovatel/todo-learning/internal/models"
	"github.com/polzovatel/todo-learning/internal/service"
	"github.com/polzovatel/todo-learning/logger"
)

type TodoController struct {
	service   service.TodoService
	jwtSigner *auth2.JWTSigner
	logger    *slog.Logger
}

func NewTodoController(service service.TodoService, jwtSigner *auth2.JWTSigner, logger *slog.Logger) *TodoController {
	return &TodoController{
		service:   service,
		jwtSigner: jwtSigner,
		logger:    logger,
	}
}

func (c *TodoController) CreateTodo(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userIDstr, exists := ctx.Get("user_id")
	if !exists {
		appLogger.Warn("user id not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDstr.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reqTodo models.CreateTodoRequest
	if err := ctx.ShouldBind(&reqTodo); err != nil {
		appLogger.Warn("invalid todo payload", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := c.service.CreateTodo(ctx, userID, reqTodo.Title, reqTodo.Description)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			appLogger.Warn("user not found while creating todo", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		appLogger.Error("failed to create todo", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	appLogger.Info("todo created", slog.String("todo_id", todo.ID.String()))
	ctx.JSON(http.StatusCreated, gin.H{"todo": todo})
}

func (c *TodoController) GetTodoByID(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userIDstr, exists := ctx.Get("user_id")
	if !exists {
		appLogger.Warn("user id not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDstr.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todoID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		appLogger.Warn("invalid todo id param", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo, err := c.service.GetTodoByID(ctx, todoID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTodoNotFound):
			appLogger.Warn("todo not found", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case errors.Is(err, domain.ErrForbidden):
			appLogger.Warn("access forbidden for todo", slog.Any("todo_id", todoID.String()))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		default:
			appLogger.Error("failed to get todo", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":          todo.ID,
		"user_id":     todo.UserID,
		"title":       todo.Title,
		"description": todo.Description,
		"completed":   todo.Completed,
		"created_at":  todo.CreatedAt,
		"updated_at":  todo.UpdatedAt,
	})
}

func (c *TodoController) GetTodos(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userIDstr, exists := ctx.Get("user_id")
	if !exists {
		appLogger.Warn("user id not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDstr.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todos, err := c.service.GetTodoByUserID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			appLogger.Warn("user not found while listing todos", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			appLogger.Error("failed to list todos", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"todos": todos})
}

func (c *TodoController) UpdateTodo(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userIDstr, exists := ctx.Get("user_id")
	if !exists {
		appLogger.Warn("user id not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDstr.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todoID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		appLogger.Warn("invalid todo id param", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.UpdateTodoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		appLogger.Warn("invalid update todo payload", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := c.service.UpdateTodo(ctx, todoID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			appLogger.Warn("update forbidden for todo", slog.Any("todo_id", todoID.String()))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case errors.Is(err, domain.ErrTodoNotFound):
			appLogger.Warn("todo not found for update", slog.Any("todo_id", todoID.String()))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			appLogger.Error("failed to update todo", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"todo": todo})
}

func (c *TodoController) DeleteTodo(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userIDstr, exists := ctx.Get("user_id")
	if !exists {
		appLogger.Warn("user id not found in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDstr.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todoID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		appLogger.Warn("invalid todo id param", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.DeleteTodo(ctx, todoID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			appLogger.Warn("delete forbidden for todo", slog.Any("todo_id", todoID.String()))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case errors.Is(err, domain.ErrTodoNotFound):
			appLogger.Warn("todo not found for delete", slog.Any("todo_id", todoID.String()))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			appLogger.Error("failed to delete todo", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	appLogger.Info("todo deleted", slog.String("todo_id", todoID.String()))
	ctx.JSON(http.StatusOK, gin.H{"message": "todo successfully deleted"})
}
