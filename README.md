# Subscriptions Service

REST-сервис для учета подписок пользователей с CRUD-операциями и агрегацией стоимости за период.

Сервис реализован как stateless API-приложение с PostgreSQL в качестве persistence-слоя и может запускаться локально, в CI и в контейнерной среде.

## Production-Oriented Summary

- Язык: `Go 1.26.1`
- HTTP: `Echo v4`
- База данных: `PostgreSQL 16`
- Data access: `pgx v5`, `squirrel`
- API docs: `Swagger (swaggo)`
- Логирование: `slog` JSON
- Миграции: SQL-файлы в `migrations/`

## Service Contract

Базовый префикс API: `/api/v1`

- `POST /subscriptions` — создать подписку
- `GET /subscriptions/:id` — получить подписку по ID
- `GET /subscriptions` — получить список с фильтрами
- `PUT /subscriptions/:id` — обновить подписку
- `DELETE /subscriptions/:id` — удалить подписку
- `GET /subscriptions/sum` — получить сумму подписок за период

Swagger UI: `http://localhost:1323/swagger/index.html`

## Domain Rules

- Даты `start_date` и `end_date` передаются в формате `MM-YYYY`.
- `end_date` — опциональное поле.
- Агрегация `sum` поддерживает фильтры `user_id`, `service_name`, `from`, `to`.
- `id` и `user_id` — UUID.

## Architecture

Слои:
- `internal/http/handler` — HTTP transport (Echo handlers).
- `internal/service` — бизнес-логика use-case уровня.
- `internal/repository` — доступ к данным (PostgreSQL).
- `internal/application` — composition root (container, wiring, server startup).

## Configuration

Конфигурация читается из `.env`. Шаблон: `samples/.env.sample`.

Ключевые переменные:
- `SERVER_PORT` — порт HTTP-сервера (default `1323`)
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_SSLMODE`
- `POSTGRES_CONNECT_TIMEOUT`

## Local Runbook

Требования:
- Go toolchain
- Docker + Docker Compose

Первичный запуск:

```bash
make setup
make run
```

`make setup`:
- создает `.env` и `docker-compose.yaml` из `samples/` (если отсутствуют)
- скачивает зависимости
- поднимает PostgreSQL
- применяет миграции

Проверка health-like ответа:

```bash
curl http://localhost:1323/api/v1/subscriptions
```

## Operational Commands

- `make help` — список команд
- `make run` — запуск API
- `make build` — сборка бинаря в `bin/`
- `make docker-up` / `make docker-down` — управление PostgreSQL
- `make migrate` — применение миграций
- `make swagger` — генерация Swagger
- `make open-swagger` — открыть Swagger UI

## Testing

Unit/интеграционные тесты:

```bash
make test
make test-integration
```

Моки:

```bash
make mocks
```

- Генерация: `mockery` по конфигу `api/mockery.yaml`
- Сгенерированные моки: `mocks/`
- Интеграционные тесты: `tests/`

## Request/Response Examples

Создание подписки:

```bash
curl -X POST http://localhost:1323/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Расчет суммы:

```bash
curl "http://localhost:1323/api/v1/subscriptions/sum?service_name=Yandex%20Plus&from=01-2025&to=12-2025"
```

## Repository Layout

```text
cmd/
  migrations/         # migration runner
  subscriptions-api/  # API entrypoint
internal/
  application/        # wiring, server lifecycle
  client/postgres/    # pgx pool init
  domain/             # entities and filters
  http/handler/       # HTTP layer
  repository/         # SQL access layer
  service/            # business logic
api/                  # tooling configs (mockery)
docs/                 # generated swagger files
migrations/           # SQL migrations
mocks/                # generated mocks
tests/                # integration tests
samples/              # env/docker-compose templates
```