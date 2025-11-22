package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polzovatel/todo-learning/cmd/app"
	"github.com/polzovatel/todo-learning/config"
	"github.com/polzovatel/todo-learning/internal/auth"
	"github.com/polzovatel/todo-learning/internal/repository/in_memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*httptest.Server, *in_memory.InMemoryRepository) {
	// Настраиваем тестовое окружение
	gin.SetMode(gin.TestMode)

	// Создаем in-memory репозиторий
	repo := in_memory.NewInMemoryRepository(slog.Default())

	// Создаем JWT signer
	cfg := &config.Config{
		JWTAlg:     "HS256",
		JWTSecret:  "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	}
	signer, err := auth.NewJWTSigner(cfg)
	require.NoError(t, err)

	// Создаем приложение
	app := app.NewApp(slog.Default(), repo, repo, nil, signer)

	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(app.Router)

	return server, repo
}

func TestRegisterAndLogin(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	client := server.Client()

	t.Run("register user successfully", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "Test123!",
		}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := client.Post(server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotNil(t, result["user"])
	})

	t.Run("login with valid credentials", func(t *testing.T) {
		// Сначала регистрируем
		reqBody := map[string]string{
			"email":    "login@example.com",
			"password": "Test123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		client.Post(server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(jsonBody))

		// Теперь логинимся
		resp, err := client.Post(server.URL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result["accessToken"])
		assert.NotEmpty(t, result["refreshToken"])
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "WrongPassword",
		}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := client.Post(server.URL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
