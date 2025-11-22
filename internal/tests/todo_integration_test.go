package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoLifecycle(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	client := server.Client()

	// Helper функция для получения токена
	getToken := func(t *testing.T) string {
		// Регистрируем и логинимся
		registerBody := map[string]string{
			"email":    "todo@example.com",
			"password": "Test123!",
		}
		jsonBody, _ := json.Marshal(registerBody)
		client.Post(server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(jsonBody))

		resp, _ := client.Post(server.URL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
		var loginResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResult)
		resp.Body.Close()

		return loginResult["accessToken"].(string)
	}

	token := getToken(t)

	t.Run("create_todo", func(t *testing.T) {
		reqBody := map[string]string{
			"title":       "Test Todo",
			"description": "Test Description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", server.URL+"/api/v1/todos", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.NotNil(t, result["todo"])
	})

	t.Run("get todos list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/v1/todos", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		todos, ok := result["todos"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, todos)
	})

	t.Run("get todo without auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/v1/todos", nil)
		// Не добавляем Authorization header

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
