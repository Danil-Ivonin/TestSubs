# TestSubs

REST-сервис для учета онлайн-подписок пользователей и расчета суммарной стоимости подписок за выбранный период.

## Возможности

- CRUDL операций над подписками.
- Расчет суммарной стоимости подписок за период.
- Фильтрация по `user_id` и `service_name`.
- PostgreSQL в качестве хранилища.
- SQL-миграции через `golang-migrate`.
- Swagger-документация.
- Запуск через Docker Compose.
- Логирование HTTP-запросов и ошибок через `logrus`.

## Стек

- Go
- Gin
- Logrus
- Viper
- godotenv
- pgx
- golang-migrate
- PostgreSQL
- Swagger / swaggo
- Docker Compose

## Конфигурация

Сервис читает `.env`, переменные окружения и опционально `config.yaml`.

Пример `.env`:

```env
HTTP_PORT=8080
HTTP_READ_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=10s
HTTP_IDLE_TIMEOUT=60s

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=subs_db
POSTGRES_SSLMODE=disable

LOG_LEVEL=info
LOG_FORMAT=text
```

Итоговая PostgreSQL DSN собирается приложением из `POSTGRES_*`.

## Запуск Через Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

Сервисы:

- `postgres` - база данных.
- `migrate` - применяет миграции.
- `app` - HTTP API.

API будет доступен на:

```text
http://localhost:8080
```

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Остановить сервисы:

```bash
docker compose down
```

## Локальный Запуск

Для локального запуска нужна доступная PostgreSQL и примененные миграции.

```bash
cp .env.example .env
go run ./cmd/subscriptions
```

Если приложение запускается вне Docker Compose, укажите `POSTGRES_HOST=localhost`.

## Миграции

Миграции лежат в `migrations`.

Через Docker Compose:

```bash
make migrate-up
```

Откат последней миграции:

```bash
make migrate-down
```

## Makefile

```bash
make test
make build
make run
make docker-build
make docker-up
make docker-down
make migrate-up
make migrate-down
make load-test
make load-test-local
```

## Нагрузочные тесты

Нагрузочный тест написан на k6 и лежит в `tests/load/subscriptions.js`.

Сценарий на каждой итерации выполняет:

- `POST /api/v1/subscriptions`
- `GET /api/v1/subscriptions/{id}`
- `GET /api/v1/subscriptions`
- `GET /api/v1/subscriptions/total`
- `PUT /api/v1/subscriptions/{id}`
- `DELETE /api/v1/subscriptions/{id}`

Перед запуском поднимите сервис:

```bash
make docker-up
```

В другом терминале запустите нагрузочный тест через Docker-образ k6:

```bash
make load-test
```

По умолчанию используется `10` виртуальных пользователей в течение `1m`.

Параметры можно переопределить:

```bash
make load-test VUS=50 DURATION=3m
```

Если API запущен не на `localhost:8080`, передайте другой адрес:

```bash
make load-test LOAD_BASE_URL=http://host.docker.internal:8081
```

Если k6 установлен локально, можно запускать без Docker:

```bash
make load-test-local
```

Для локального k6 адрес задается отдельно:

```bash
make load-test-local LOCAL_LOAD_BASE_URL=http://localhost:8080 VUS=20 DURATION=2m
```

## API

Базовый путь:

```text
/api/v1
```

### Создать Подписку

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Ответ:

```json
{
  "id": "f3a9c5a8-9a8b-4a3e-9f15-5d5b4d6c3f52",
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

### Получить Подписку

```bash
curl http://localhost:8080/api/v1/subscriptions/{id}
```

### Список Подписок

```bash
curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&limit=20&offset=0"
```

### Обновить Подписку

```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 500,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

### Удалить Подписку

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/{id}
```

### Рассчитать Сумму

Период задается включительно в формате `MM-YYYY`.

```bash
curl "http://localhost:8080/api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

Ответ:

```json
{
  "total": 2400
}
```

## Формат Дат

API принимает даты подписок в формате:

```text
MM-YYYY
```

В PostgreSQL даты хранятся как `date`, нормализованные к первому дню месяца.

## Проверка

```bash
go test ./...
go build ./cmd/subscriptions
docker compose config
docker compose build
```
