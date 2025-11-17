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

type UserController struct {
	service   service.Service
	jwtSigner *auth2.JWTSigner
	logger    *slog.Logger
}

func NewUserController(service service.Service, jwtSigner *auth2.JWTSigner, logger *slog.Logger) *UserController {
	return &UserController{
		service:   service,
		jwtSigner: jwtSigner,
		logger:    logger,
	}
}

func (c *UserController) RegisterUser(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	var req models.RegisterRequest

	// Parse JSON
	if err := ctx.ShouldBind(&req); err != nil {
		appLogger.Warn("invalid register payload", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, err := auth2.HashPassword(req.Password)
	if err != nil {
		appLogger.Error("failed to hash password", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.CreateUser(ctx, req.Email, hash)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			appLogger.Warn("email already taken", slog.String("email", req.Email))
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": domain.ErrEmailTaken.Error()})
			return
		}
		appLogger.Error("failed to create user", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	appLogger.Info("user created", slog.String("email", user.Email))
	ctx.JSON(http.StatusCreated, gin.H{"user": user})
}

func (c *UserController) LoginUser(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	var req models.LoginRequest

	if err := ctx.ShouldBind(&req); err != nil {
		appLogger.Warn("invalid login payload", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			appLogger.Warn("user not found for login", slog.String("email", req.Email))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		default:
			appLogger.Error("failed to fetch user by email", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	hashTrue, err := auth2.CheckPasswordHash(req.Password, user.PasswordHash)
	if err != nil {
		appLogger.Error("failed to compare password hash", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !hashTrue {
		appLogger.Warn("invalid credentials")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accessToken, err := c.jwtSigner.GenerateAccessToken(user.ID.String(), user.Email, "access_token")
	if err != nil {
		appLogger.Error("failed to generate access token", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := c.jwtSigner.GenerateRefreshToken(user.ID.String(), user.Email, "refresh_token")
	if err != nil {
		appLogger.Error("failed to generate refresh token", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	appLogger.Info("user logged in", slog.String("email", user.Email))
	ctx.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}

func (c *UserController) LogoutUser(ctx *gin.Context) {
	// Просто подтверждаем выход
	// В реальном приложении здесь можно добавить blacklist токенов
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (c *UserController) GetMe(ctx *gin.Context) {
	appLogger := logger.LoggerFromContext(ctx, c.logger)
	userID, exist := ctx.Get("user_id")
	if !exist {
		appLogger.Warn("user id missing in context")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDParsed, err := uuid.Parse(userID.(string))
	if err != nil {
		appLogger.Error("failed to parse user id", slog.Any("error", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.GetUserById(ctx, userIDParsed)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			appLogger.Warn("user not found in storage", slog.String("user_id", userIDParsed.String()))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": domain.ErrUserNotFound.Error()})
			return
		default:
			appLogger.Error("failed to fetch user", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"createdAt": user.CreatedAt,
	})
}
