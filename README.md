# my-todo-learning

Учебный todo-сервис на Go 1.21+. Проект используется как тренажёр: REST API, JWT-аутентификация, PostgreSQL с fallback на in-memory хранилище, структурное логирование и базовые дев-команды.

## Требования

- Go 1.21+
- Docker + Docker Compose

## Быстрый старт

```bash
# поднять инфраструктуру (Postgres)
make up

# запустить приложение (по умолчанию :8080)
make run

# остановить окружение и удалить тома
make down
```

Во время старта сервис пробует подключиться к Postgres. Если соединение или миграции не проходят, приложение логирует предупреждение и работает с in-memory репозиторием (данные пропадают при перезапуске).

## Полезные команды

```bash
make build   # собрать бинарник в bin/my-todo-learning
make run     # go run ./cmd/main.go
make test    # go test ./...
make fmt     # go fmt ./...
make lint    # go vet ./...
make clean   # удалить bin/, coverage.*
```

## Структура каталога

```
cmd/
  main.go               точка входа, DI, регистрация маршрутов
  middleware/           HTTP middleware (auth, request logging)
config/                 загрузка конфигурации из env
internal/
  auth/                 JWT и bcrypt
  controller/           HTTP-обработчики (Gin)
  database/             pgx pool + миграции
  domain/               доменные ошибки
  models/               структуры данных и DTO
  repository/           реализации хранилищ (postgres, in-memory)
  service/              бизнес-логика users/todos
logger/                 настройка slog
```

## REST API

Все маршруты находятся под префиксом `/api/v1`.

### Публичные
- `POST /register` — регистрация пользователя (email + пароль)
- `POST /login` — вход, возвращает access/refresh JWT

### Защищённые (требуется `Authorization: Bearer <token>`)
- `GET /me` — профиль текущего пользователя
- `POST /logout` — заглушка для выхода
- `POST /todos` — создать задачу
- `GET /todos` — список задач пользователя
- `GET /todos/:id` — получить задачу
- `PUT /todos/:id` — обновить задачу
- `DELETE /todos/:id` — удалить задачу

## Логи и мониторинг

- Используется `log/slog` (Go 1.21) + собственный middleware, который добавляет `request_id`, HTTP-метод и путь.
- Репозитории и сервисы пишут ключевые события (`Warn` при отсутствии данных, `Error` при сбоях, `Info` при успешных операциях).

## Тесты

Интеграционные тесты планируются в пакете `internal/tests` (Day 4, шаг 3). На данный момент можно запускать юнит-тесты:

```bash
make test
```

## Документация по обучению

Материалы с пошаговым планом лежат в родительской директории репозитория:
- `LEARNING_PLAN_GOLANG_FIXED.md`
- `DAY1_STEP_BY_STEP.md`, `DAY2_STEP_BY_STEP.md`, `DAY3_STEP_BY_STEP.md`, `DAY4_STEP_BY_STEP.md`

## Лицензия

Учебный проект. Используйте как шаблон и дорабатывайте под свои задачи.***
