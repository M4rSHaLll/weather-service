# Weather service

Сервис раз в 30 минут получает координаты города из Open-Meteo, запрашивает текущую
температуру, сохраняет измерение в PostgreSQL, кэширует его в Redis и возвращает последнее измерение по
`GET /{city}`. Если города ещё нет в базе, данные запрашиваются у Open-Meteo и
сохраняются непосредственно во время HTTP-запроса.

## Запуск

```bash
docker compose up -d
go run ./cmd/api
```

Пример запроса:

```bash
curl http://localhost:3000/moscow
```

Перед первым запуском создайте локальный файл окружения:

```bash
cp .env.example .env
```

На Windows PowerShell: `Copy-Item .env.example .env`.

## Конфигурация

| Переменная | Значение по умолчанию |
|---|---|
| `HTTP_ADDR` | `:3000` |
| `DATABASE_URL` | обязательно, задаётся в `.env` |
| `POSTGRES_PORT` | `15432` |
| `REDIS_URL` | обязательно, задаётся в `.env` |
| `REDIS_PORT` | `16379` |
| `CACHE_TTL` | `2m` |
| `REDIS_TIMEOUT` | `500ms` |
| `WEATHER_CITY` | `moscow` |
| `WEATHER_POLL_INTERVAL` | `30m` |
| `HTTP_CLIENT_TIMEOUT` | `10s` |
| `SHUTDOWN_TIMEOUT` | `10s` |

Миграция из `migrations/` автоматически применяется только при создании нового
тома PostgreSQL. Для уже существующей базы её нужно выполнить отдельно.

При чтении сервис сначала обращается к Redis. Промах кэша заполняется данными из
PostgreSQL, а новые измерения записываются в Redis после успешной записи в базу.
Недоступность Redis не блокирует работу сервиса: источником истины остаётся PostgreSQL.
