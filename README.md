# CRM API Gateway

API Gateway для CRM-сервиса на Go. Единая точка входа для клиентов: маршрутизация HTTP-запросов к внутренним gRPC-сервисам (CRM и Auth), JWT-аутентификация, кэширование в Redis и структурированное логирование.

## О проекте

Шлюз принимает внешние HTTP-запросы и проксирует их по gRPC в два внутренних сервиса:

- **CRM** (`CRM_URL`) — операции с бронированиями
- **Auth** (`AUTH_URL`) — аутентификация, регистрация, refresh/выход

На уровне шлюза выполняются:

- Проверка JWT через `middleware.AuthMiddleware` (делегирует в `authService`)
- Кэширование ответов Auth-сервиса в Redis
- Единый формат ошибок и логирование через `logrus`

## Технологический стек

| Компонент | Технология |
|---|---|
| Язык | Go |
| HTTP-фреймворк | [Gin](https://github.com/gin-gonic/gin) |
| Межсервисное взаимодействие | gRPC (`google.golang.org/grpc`) |
| Кэш | Redis (`pkg/cache`) |
| Конфигурация | `.env` через `godotenv` |
| Логирование | `sirupsen/logrus` (уровень Debug) |

## Быстрый старт

### Требования

- Go 1.21+
- Запущенные gRPC-сервисы CRM и Auth
- Redis

### Клонирование и сборка

```bash
git clone https://github.com/JanMird/gateway.git
cd gateway

go mod download
go build -o gateway .
```

### Переменные окружения

Создайте `.env` в корне проекта:

```env
# Redis
REDIS_ENDPOINT=localhost:6379
RESID_USER=
REDIS_PASSWORD=
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5
REDIS_TIMEOUT=3

# gRPC-бэкенды
CRM_URL=localhost:50051
AUTH_URL=localhost:50052
```

### Запуск

```bash
go run .
# или
./gateway
```

Сервер слушает `:8081` (порт захардкожен в `main.go`).

## HTTP API

### Публичные эндпоинты (`/auth`)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/auth/authenticate` | Аутентификация пользователя |
| POST | `/auth/signup` | Регистрация |
| POST | `/auth/refresh` | Обновление access-токена |
| GET | `/auth/health` | Проверка живости Auth-сервиса |
| POST | `/auth/signout` | Выход |

### Защищённые эндпоинты (`/api`, требуется JWT)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/book/booking` | Создание бронирования |

Защищённые маршруты обёрнуты в `middleware.AuthMiddleware(authService)`. Ожидается заголовок:

```
Authorization: Bearer <JWT>
```

## Архитектура

```
HTTP (Gin) → handler → service → gRPC client → backend (CRM / Auth)
                    ↘ cache (Redis)
```

Основные слои:

- `internal/handler` — HTTP-хендлеры (`AuthHandler`, `BookingHandler`)
- `internal/service` — бизнес-логика (`AuthService`, `BookingService`), опции `WithCache`, `WithLogger`
- `internal/client` — gRPC-клиенты (`GRPCAuthClient`, `GRPCBookingClient`)
- `internal/middleware` — `AuthMiddleware` для защиты `/api`
- `internal/domain` — доменные типы и константы (например, `GRPCTimeout`)
- `pkg/cache` — обёртка над Redis (`Config`, `NewRedisClient`, `New`)

## Структура проекта

```
gateway/
├── cmd/
|   ├── main.go             # точка входа: конфиг, gRPC-коннекты, роутинг
├── gen/
|   ├── auth/
|   ├── booking/
├── internal/
│   ├── client/             # gRPC-клиенты
│   ├── domain/             # доменные типы, таймауты
│   ├── handler/            # HTTP-хендлеры
│   ├── middleware/         # AuthMiddleware
│   └── service/            # бизнес-логика + опции (cache, logger)
├── pkg/cache/              # Redis-клиент
└── .env
```

## Тестирование

```bash
go test ./...

# Ручная проверка
curl -i http://localhost:8081/auth/health

curl -i -X POST http://localhost:8081/api/book/booking \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"...":"..."}'
```
