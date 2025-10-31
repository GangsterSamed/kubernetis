# 📚 Todo Learning Project

## 🎯 Цель проекта

Учебный проект для изучения Golang и технологий за 2 недели.

## 📋 План обучения

- **День 1-2:** Go основы + REST API + JWT
- **День 3-4:** PostgreSQL + pgx + миграции  
- **День 5-6:** Clean architecture + паттерны
- **День 7:** Testing + Git workflow
- **День 8-9:** Docker + Redis + workers
- **День 10-11:** gRPC + Protocol Buffers
- **День 12:** Kafka basics
- **День 13:** Kubernetes deployment
- **День 14:** Мониторинг + финализация

## 🚀 Начало работы

### День 1

```bash
# Установите зависимости
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid

# Запустите сервер
go run main.go
```

## 📖 Документация

Смотрите файлы в родительской папке:
- `START_HERE_GOLANG.md` - точка входа
- `LEARNING_PLAN_GOLANG.md` - подробный план
- `CODE_PATTERNS_GOLANG.md` - шпаргалка

## 🏗️ Структура проекта

```
my-todo-learning/
├── main.go              # Entry point
├── handlers.go          # HTTP handlers
├── models.go            # Data models
├── auth.go              # JWT & passwords
├── repo.go              # Repository pattern
├── go.mod
├── go.sum
└── README.md
```

## 🧪 Тестирование

```bash
go test ./...
go test -v
```

## 📝 Commands

```bash
# Run
go run .

# Build
go build -o bin/todo-api

# Format
go fmt ./...

# Lint
go vet ./...

# Test
go test ./...
```

## 🐳 Docker

```bash
docker build -t todo-api .
docker run -p 8080:8080 todo-api
```

## 📦 Docker Compose

```bash
docker-compose up
```

## 🔗 API Endpoints

### Public
- `POST /api/v1/register` - Регистрация
- `POST /api/v1/login` - Вход
- `GET /health` - Health check

### Protected
- `GET /api/v1/me` - Текущий пользователь
- `GET /api/v1/todos` - Список задач
- `POST /api/v1/todos` - Создать задачу
- `GET /api/v1/todos/:id` - Получить задачу
- `PATCH /api/v1/todos/:id` - Обновить задачу
- `DELETE /api/v1/todos/:id` - Удалить задачу

## 📚 Resources

- [Go Tour](https://go.dev/tour/)
- [Go by Example](https://gobyexample.com/)
- [pgx docs](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [gRPC Go](https://grpc.io/docs/languages/go/)

## 🎯 Progress Tracker

- [ ] День 1: REST API + JWT
- [ ] День 2: Завершить REST API
- [ ] День 3: PostgreSQL интеграция
- [ ] День 4: Миграции
- [ ] День 5: Clean architecture
- [ ] День 6: Repository pattern
- [ ] День 7: Тесты
- [ ] День 8: Docker
- [ ] День 9: Redis
- [ ] День 10: gRPC
- [ ] День 11: gRPC client/server
- [ ] День 12: Kafka
- [ ] День 13: Kubernetes
- [ ] День 14: Мониторинг

---

**Следуйте плану и учитесь! 💪**
