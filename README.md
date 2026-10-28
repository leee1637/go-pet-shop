# go-pet-shop

Учебный pet-проект интернет-магазина для питомцев на Go.

## Стек
- **Go 1.25** — язык
- **Gin** — HTTP роутер
- **PostgreSQL 16** + **pgx/v5** — БД и драйвер
- **golang-migrate** — миграции
- **cleanenv** + **godotenv** — конфиг
- **slog** — логирование
- **Docker + docker-compose** — поднятие всего стека одной командой

## Структура проекта
```text
.
├── cmd/app        # Точка входа HTTP-сервера
├── cmd/migrator   # Утилита для применения миграций
├── config         # YAML-конфиги по окружениям (local.yaml, dev.yaml)
├── internal
│   ├── config     # Загрузка конфига
│   ├── handlers   # HTTP-хендлеры + тесты
│   ├── lib/logger # Настройка логгера
│   ├── models     # Доменные модели
│   └── storage
│       └── postgres # Реализация репозитория
├── migrations     # SQL-миграции (up/down)
└── docker-compose.yml
```

## Что реализовано / добавлено для тренировки
- CRUD для пользователей, товаров, заказов
- `POST /checkout` — оформление заказа в транзакции (`PlaceOrder`)
- `GET /users/:email/history` — история заказов с позициями
- `GET /products/popular` — популярные товары по количеству продаж
- Unit-тесты хендлеров (`product`, `user`, `order`) с моками
- Graceful shutdown
- Логирование через `slog`
- Автоприменение миграций при поднятии через docker-compose

## Быстрый запуск (одной командой)

1) Скопируй `.env.example` в `.env`
```bash
cp .env.example .env
```

2) Подними весь стек
```bash
docker compose up --build
```

При старте `migrator` применит все миграции из `./migrations`, после успешного завершения запустится `app`. Postgres будет готов по healthcheck.

Проверка:
```bash
curl http://localhost:8080/health
```

Остановка:
```bash
docker compose down
```
С удалением тома БД (чистый старт):
```bash
docker compose down -v
```

## Локальный запуск без Docker (опционально)

Если хочешь запустить только сервер локально, а БД в Docker:

```bash
# Поднять только Postgres
docker compose up -d postgres

# Скопировать .env под локальный запуск
cp .env.example .env
# Для локального запуска DATABASE_URL можно поменять на: postgres://user:password@localhost:5432/petshop?sslmode=disable
# APP_ENV=local

# Накатить миграции (отдельно)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path ./migrations -database "postgres://user:password@localhost:5432/petshop?sslmode=disable" up
# либо: go run ./cmd/migrator --migrations-path ./migrations

# Запустить сервер
go run ./cmd/app
```

## Основные эндпоинты
| Метод | URL | Описание |
|---|---|---|
| GET | `/health` | Проверка состояния |
| GET | `/users` | Список пользователей |
| GET | `/users/:email` | Пользователь по email |
| POST | `/users` | Создать пользователя |
| GET | `/users/:email/orders` | Заказы пользователя |
| GET | `/users/:email/history` | История заказов и транзакций |
| GET | `/products` | Список товаров |
| POST | `/products` | Добавить товар |
| GET | `/products/:id` | Товар по ID |
| PUT | `/products/:id` | Обновить товар |
| DELETE | `/products/:id` | Удалить товар |
| GET | `/products/popular` | Популярные товары |
| POST | `/orders` | Создать заказ |
| GET | `/orders/:id` | Детали заказа |
| POST | `/orders/:id/items` | Добавить позицию к заказу |
| POST | `/checkout` | Оформление заказа (транзакция) |

> Точное # PR Title
<!-- PR title must follow conventional commit format -->
# refactor: migrate setup to docker-compose with auto migrations

## Description
<!-- Describe your changes in detail -->
This PR introduces a unified docker-compose setup where migrations run automatically on startup via a dedicated `migrator` service. Also adds dev config, updates Dockerfile, and improves README/env examples for one-command startup.

## Changes
- Added `config/dev.yaml`
- Updated `.env.example` for Postgres healthcheck and docker networking
- Reworked `docker-compose.yml`: added `migrator` service with `service_completed_successfully` dependency
- Updated `Dockerfile` to build both binaries and run app without inline migration command
- Updated `README.md` with quick start (`docker compose up --build`) and notes on local run
- Added/updated unit tests for handlers with mocks

## How to test
```bash
cp .env.example .env
docker compose up --build
curl http://localhost:8080/health
```
```bash
go test ./internal/handlers/... -v
go test ./... -cover
```

## Checklist
- [x] Tests pass
- [x] README updated
- [x] Docker setup works with auto migrations

Соответствие смотри в `cmd/app/main.go`

## Тесты
```bash
# Все тесты
go test ./...

# Тесты хендлеров с подробным выводом
go test ./internal/handlers/... -v

# Конкретный тест
go test ./internal/handlers/order -v -run TestPlaceOrder

# Покрытие по пакетам
go test ./... -cover

# HTML-отчёт покрытия
go test ./internal/handlers/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Покрытие по функциям
go tool cover -func=coverage.out
```

## План по дальнейшей тренировке
- **Метрики** — Prometheus + Grafana, middleware для счётчиков/гистограмм
- **gRPC** — proto + gRPC-сервер, можно запустить параллельно с HTTP
- **Валидация** — `go-playground/validator`
- **Интеграционные тесты** — `testcontainers-go`
- **CI** — GitHub Actions (`go vet`, `golangci-lint`, `go test -race`)
