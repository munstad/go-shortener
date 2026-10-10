# go-shortener

Простой сервис сокращения ссылок на Go. Без внешних зависимостей, только стандартная библиотека.

## Возможности

- Создание коротких ссылок через REST API
- Редирект по короткому коду со счетчиком переходов
- Хранение в JSON-файле (атомарная запись), потокобезопасное хранилище
- Graceful shutdown, healthcheck, Dockerfile, тесты

## Запуск

```bash
go run . -addr :8080 -data links.json
```

Или через Docker:

```bash
docker build -t go-shortener .
docker run -p 8080:8080 -v shortener-data:/app/data go-shortener
```

## API

Создать ссылку:

```bash
curl -X POST localhost:8080/api/links -d '{"url":"https://go.dev"}'
# {"code":"aB3xYz","short_url":"http://localhost:8080/aB3xYz"}
```

Перейти по ссылке: `GET /{code}` (редирект 302).

Статистика: `GET /api/links/{code}`

```json
{"code":"aB3xYz","url":"https://go.dev","clicks":2,"created_at":"2026-10-10T12:00:00Z"}
```

Healthcheck: `GET /healthz`

## Тесты

```bash
go test ./...
```

## Идеи для развития

- Свои алиасы и срок жизни ссылок
- Хранилище на SQLite или PostgreSQL
- Rate limiting и метрики Prometheus
- CI на GitHub Actions
