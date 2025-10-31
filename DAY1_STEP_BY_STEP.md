# 📝 День 1: Пошаговая инструкция

## 🎯 Цель дня
Создать минимальный REST API с регистрацией, логином и защищенными endpoint'ами

---

## Шаг 1: Структура данных (30 минут)

### Что нужно сделать:

**1.1.** Создайте файл `models.go`  
**1.2.** Определите структуру User:

Подсказки:
- Используйте UUID для ID (библиотека `google/uuid`)
- Email, пароль
- Timestamp для `created_at`

Формат подсказки:
```go
type User struct {
    ID           (что-то для уникальности) 
    Email        (текст)
    PasswordHash (текст - не plain password!)
    CreatedAt    (что-то для времени)
}
```

**1.3.** Определите структуру для регистрации/логина:

Вам нужны структуры для request'ов:
- Одна для регистрации (поля: email, password)
- Одна для логина (поля: email, password)

---

## Шаг 2: In-memory хранилище (45 минут)

### Что нужно сделать:

**2.1.** Создайте файл `store.go`  
**2.2.** Создайте структуру UserStore:

Подсказка структуры:
```go
type UserStore struct {
    users map[...тип ключа...] *...тип значения...
}
```

**2.3.** Добавьте методы:

Напишите функции для:
- `Create(email, passwordHash)` - создать пользователя
- `GetByEmail(email)` - найти по email
- `GetByID(id)` - найти по ID

Для каждого метода:
- Обрабатывайте ошибки (что если email уже занят?)
- Возвращайте указатель на User или nil

---

## Шаг 3: Password hashing (30 минут)

### Что нужно сделать:

**3.1.** Создайте файл `auth.go`  
**3.2.** Установите зависимость:
```bash
go get golang.org/x/crypto/bcrypt
```

**3.3.** Напишите две функции:

**Функция 1:** HashPassword
- Принимает: строку (plain password)
- Возвращает: строку (hash) и error
- Использует: `bcrypt.GenerateFromPassword`

**Функция 2:** CheckPassword
- Принимает: две строки (plain password, hash)
- Возвращает: bool
- Использует: `bcrypt.CompareHashAndPassword`

---

## Шаг 4: JWT токены (1 час)

### Что нужно сделать:

**4.1.** Установите JWT библиотеку:
```bash
go get github.com/golang-jwt/jwt/v5
```

**4.2.** Создайте структуру для claims:
```go
type Claims struct {
    UserID (UUID или string)
    Email  (string)
    jwt.RegisteredClaims
}
```

**4.3.** Напишите функцию GenerateAccessToken:

Алгоритм:
1. Создайте Claims со временем жизни (например, 30 минут)
2. Создайте токен с методом HS256
3. Подпишите токен секретным ключом
4. Верните строку токена и error

Константы:
```go
jwtSecret = "your-secret-key-change-later"
accessTTL = 30 * time.Minute
```

**4.4.** Напишите функцию ValidateToken:

Алгоритм:
1. Распарсите токен с claims
2. Проверьте валидность
3. Верните Claims или error

---

## Шаг 5: HTTP Handlers - Регистрация (45 минут)

### Что нужно сделать:

**5.1.** Создайте файл `handlers.go`  
**5.2.** Структура AuthHandler:

```go
type AuthHandler struct {
    store (ваш UserStore)
}
```

**5.3.** Метод Register:

Что должно происходить:
1. Распарсить JSON из request body
2. Хешировать пароль
3. Создать пользователя через store
4. Вернуть пользователя (БЕЗ пароля!) в JSON
5. HTTP статус: 201 Created при успехе, 400 при ошибке

Формат подсказки для Gin:
```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req (ваша структура request)
    
    // Parse JSON
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "..."})
        return
    }
    
    // Hash password
    hash, err := HashPassword(req.Password)
    
    // Create user
    user, err := h.store.Create(...)
    
    // Return response
    c.JSON(201, gin.H{...})
}
```

---

## Шаг 6: HTTP Handlers - Логин (45 минут)

### Что нужно сделать:

**6.1.** Метод Login в том же файле:

Что должно происходить:
1. Распарсить email и password
2. Найти пользователя по email
3. Проверить пароль
4. Сгенерировать access токен
5. Вернуть токен в JSON
6. HTTP статус: 200 при успехе, 401 при неверных данных

Формат ответа:
```json
{
    "access_token": "...",
    "user": {
        "id": "...",
        "email": "..."
    }
}
```

---

## Шаг 7: Middleware для проверки токена (45 минут)

### Что нужно сделать:

**7.1.** Напишите middleware функцию:

Алгоритм:
1. Извлечь токен из заголовка Authorization
2. Формат: "Bearer <token>"
3. Валидировать токен
4. Сохранить user_id в контекст (c.Set)
5. Продолжить выполнение (c.Next()) или прервать (c.Abort())

Подсказка:
```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        
        // Извлечь токен из "Bearer <token>"
        
        // Валидировать
        
        // Сохранить user_id в контекст
        
        c.Next()
    }
}
```

---

## Шаг 8: Todos endpoints (1 час)

### Что нужно сделать:

**8.1.** Добавьте Todo структуру в models.go

**8.2.** Создайте TodoStore аналогично UserStore

**8.3.** Создайте TodoHandler:

Методы:
- List - получить все todos текущего пользователя (используйте c.Get("user_id"))
- Create - создать новый todo
- GetByID - получить один todo
- Update - обновить todo
- Delete - удалить todo

Важно: Каждый метод проверяет, что todo принадлежит текущему пользователю!

---

## Шаг 9: Собрать всё вместе (30 минут)

### Что нужно сделать:

**9.1.** В файле `main.go`:

Создайте:
1. Router (gin.Default())
2. UserStore и TodoStore
3. Handlers
4. Register публичные endpoints: /register, /login
5. Register защищенные endpoints: /todos, /me
6. Запустите сервер на порту 8080

Структура:
```go
func main() {
    r := gin.Default()
    
    // Создайте stores
    
    // Создайте handlers
    
    // Публичные роуты
    r.POST("/api/v1/register", ...)
    r.POST("/api/v1/login", ...)
    
    // Защищенные роуты
    protected := r.Group("/api/v1")
    protected.Use(authMiddleware())
    {
        protected.GET("/me", ...)
        protected.GET("/todos", ...)
        protected.POST("/todos", ...)
        // и т.д.
    }
    
    r.Run(":8080")
}
```

---

## Шаг 10: Тестирование (30 минут)

### Что нужно сделать:

**10.1.** Регистрация:
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}'
```

**10.2.** Логин (скопируйте access_token):
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}'
```

**10.3.** Создать Todo:
```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"title":"Learn Go","description":"Complete day 1"}'
```

**10.4.** Получить список Todos:
```bash
curl http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## ✅ Чеклист завершения Дня 1

- [ ] Есть структуры User и Todo
- [ ] Есть in-memory хранилища
- [ ] Работает хеширование паролей
- [ ] Работает генерация JWT токенов
- [ ] Работает валидация JWT
- [ ] Работает регистрация (/register)
- [ ] Работает логин (/login)
- [ ] Работает middleware проверки токена
- [ ] Работает CRUD для Todos
- [ ] Все протестировано через curl
- [ ] Код закоммичен в Git

---

## 🐛 Если что-то не работает

**Ошибка компиляции:**
- Проверьте импорты
- `go mod tidy` - автоматически подтянет зависимости

**Забыли как что-то сделать:**
- Посмотрите в CODE_PATTERNS_GOLANG.md
- Посмотрите в authsvc проект (но не копируйте!)
- Google: "golang gin <что нужно>"

**Токен не работает:**
- Проверьте что вы правильно извлекаете из заголовка
- Проверьте что используете тот же jwtSecret для генерации и валидации

---

## 🎯 Главное правило

**Пишите код сами!** Используйте подсказки и примеры как направление, но НЕ копируйте готовый код.

**Если застряли на каком-то шаге больше 30 минут** - переходите к следующему и возвращайтесь позже.
