# Заготовка социальной сети (домашнее задание)
## Описание

В проекте реализован минимальный набор функционала социальной сети:
- Регистрация пользователей с указанием профиля (имя, фамилия, дата рождения, пол, интересы, город).
- Авторизация с выдачей сессионного токена.
- Получение данных пользователя по его ID.

Все SQL-запросы параметризованы и обеспечивают защиту от SQL инъекций.
Пароли хранятся в безопасном хешированном виде с использованием алгоритма `bcrypt`.

Проект представляет собой монолитное приложение
на **Go** с БД **PostgreSQL** (без ORM).

## Структура проекта

```Схема
social_network/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа, настройка HTTP-роутера
├── internal/
│   ├── db/
│   │   └── db.go                # Инициализация пула соединений с БД
│   ├── handlers/
│   │   └── handlers.go          # HTTP-обработчики и middleware авторизации
│   └── models/
│       └── models.go            # Структуры данных (User, LoginRequest и др.)
├── migrations/
│   └── 001_init.sql             # SQL-скрипт инициализации базы данных
├── docker-compose.yml           # Конфигурация для локального запуска PostgreSQL
├── go.mod                       # Описание модуля Go и зависимостей
└── README.md                    # Документация проекта
```

## Требования

Для запуска проекта локально необходимо установить:
- **Go** версии 1.22 или выше ([скачать](https://go.dev/dl/))
- **Docker** и **Docker Compose** ([скачать](https://www.docker.com/products/docker-desktop/))

## Инструкция по локальному запуску

### Шаг 1. Клонирование репозитория

```bash
git clone <ссылка-на-репозиторий>
cd social_network
```

### Шаг 2. Запуск базы данных PostgreSQL
Запустить контейнер с PostgreSQL с помощью Docker Compose.
При старте автоматически применится миграция из `migrations/001_init.sql`.

```bash
docker-compose up -d
```
Проверить, что база поднялась:
```bash
docker-compose ps
```

### Шаг 3. Установка зависимостей Go
```bash
go mod tidy
```

### Шаг 4. Запуск приложения
```bash
go run cmd/app/main.go
```
При успешном старте в терминале появится сообщение:

```Результат
Successfully connected to PostgreSQL
Server started on :8080
```

Сервис будет доступен по адресу `http://localhost:8080`.

## API

Базовый URL: `http://localhost:8080`

| Метод | Эндпоинт | Описание | Авторизация |
|-------|----------|----------|-------------|
| `POST` | `/user/register` | Регистрация нового пользователя | ❌ Нет |
| `POST` | `/login` | Авторизация и получение токена | ❌ Нет |
| `GET` | `/user/get/{id}` | Получение анкеты пользователя по ID | ✅ Требуется |

### Примеры запросов

#### 1. Регистрация пользователя

```bash
curl -X POST http://localhost:8080/user/register \
-H "Content-Type: application/json" \
-d '{
  "login": "ivan_ivanov",
  "password": "super_secret_password",
  "first_name": "Иван",
  "last_name": "Иванов",
  "birth_date": "1995-05-15T00:00:00Z",
  "gender": "male",
  "interests": "программирование, спорт",
  "city": "Москва"
}'
```

**Ответ:** `{"id":1}` (статус 201)

#### 2. Авторизация

```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{
  "login": "ivan_ivanov",
  "password": "super_secret_password"
}'
```

**Ответ:** `{"token":"a1b2c3d4e5f6..."}` (статус 200)

#### 3. Получение анкеты

```bash
curl -X GET http://localhost:8080/user/get/1 \
-H "Authorization: <ВАШ_ТОКЕН>"
```

**Ответ:**

```json
{
  "id": 1,
  "login": "ivan_ivanov",
  "first_name": "Иван",
  "last_name": "Иванов",
  "birth_date": "1995-05-15T00:00:00Z",
  "gender": "male",
  "interests": "программирование, спорт",
  "city": "Москва"
}
```

## Остановка приложения

Для остановки базы данных:

```bash
docker-compose down
```

Чтобы удалить базу данных с сохраненными данными, добавить флаг `-v`:

```bash
docker-compose down -v
```