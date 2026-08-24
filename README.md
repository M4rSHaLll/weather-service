# Weather service

Сервис раз в 30 минут получает координаты Города из Open-Meteo, запрашивает текущую
температуру, сохраняет измерение в PostgreSQL и возвращает последнее измерение по
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

## Конфигурация

| Переменная | Значение по умолчанию |
|---|---|
| `HTTP_ADDR` | `:3000` |
| `DATABASE_URL` | `postgresql://postgres:a864653K@localhost:54321/weather` |
| `WEATHER_CITY` | `moscow` |
| `WEATHER_POLL_INTERVAL` | `30m` |
| `HTTP_CLIENT_TIMEOUT` | `10s` |
| `SHUTDOWN_TIMEOUT` | `10s` |

Миграция из `migrations/` автоматически применяется только при создании нового
тома PostgreSQL. Для уже существующей базы её нужно выполнить отдельно.
