package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/logger"
)

func AuthMiddleware(signer *auth.JWTSigner, appLogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqLogger := logger.LoggerFromContext(c, appLogger)
		// 1. Извлекаем токен
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			reqLogger.Warn("Authorization header is empty", slog.String("value", authHeader))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is empty"})
			return
		}

		// 2. Проверяем формат
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			reqLogger.Warn("Authorization header is invalid", slog.String("value", authHeader))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is invalid"})
			return
		}
		tokenString := parts[1]

		// 3. Валидировать токен
		claims, err := signer.ValidateToken(tokenString)
		if err != nil {
			reqLogger.Error("Token validation failed", slog.Any("error", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		// 4. Сохранить данные из токена в контекст
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("type", claims.Type)

		// 5. Передать в handler
		reqLogger.Info("token validated", slog.String("user_id", claims.UserID), slog.String("token_type", claims.Type))
		c.Next()
	}
}

func RequestLoggerMiddleware(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.NewString()

		reqLogger := base.With(
			slog.String("request_id", reqID),
			slog.String("request_method", c.Request.Method),
			slog.String("request_path", c.Request.URL.Path),
		)

		c.Set("request_id", reqID)
		c.Set("logger", reqLogger)

		reqLogger.Info("request started")
		c.Next()
		reqLogger.Info("request completed", slog.Int("status", c.Writer.Status()), slog.Duration("duration", time.Since(start)))
	}
}
