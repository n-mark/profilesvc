# Profile Service (Сервис профилей)

Микросервис профилей пользователей для интернет-сервиса по размещению объявлений (финальный проект). Написан на Go.

- **GitHub:** https://github.com/n-mark/profilesvc
- **DockerHub:** [`mblkuta/profile-service`](https://hub.docker.com/r/mblkuta/profile-service)

## Возможности

- CRUD профилей пользователей (`/api/v1/profile`)
- Публикация события `profile.created` в Kafka (топик `profile`)
- Подписка на события биллинга (`billing.account.created`) для связывания профиля с биллинг-аккаунтом
- Метрики Prometheus

## Технологии

- Go, PostgreSQL, Kafka
- Docker / docker-compose

## Структура проекта

```text
main.go        # точка входа
internal/      # бизнес-логика, обработчики, хранилище
db/init/       # SQL-инициализация БД
```

## Переменные окружения

| Переменная | Описание | Пример |
|---|---|---|
| `APP_PORT` / `SERVER_ADDR` | Порт HTTP-сервера | `8080` |
| `DB_HOST` / `DB_PORT` | Хост/порт PostgreSQL | `postgres-profiles` / `5432` |
| `DB_NAME` | Имя БД | `profiledb` |
| `DB_USER` / `DB_PASSWORD` | Учётные данные БД | `profile_user` |
| `BROKER_TYPE` | Тип брокера | `KAFKA` |
| `KAFKA_BROKERS` | Адреса брокеров Kafka | `kafka:9092` |
| `KAFKA_PROFILE_TOPIC` | Топик событий профилей | `profile` |
| `KAFKA_BILLING_TOPIC` / `KAFKA_BILLING_GROUP` | Топик/группа биллинга | `billing` / `profilesvc.billing` |
| `KAFKA_PROFILE_CREATED_EVENT_TYPE` | Тип события | `profile.created` |
| `KAFKA_BILLING_ACCOUNT_CREATED_EVENT_TYPE` | Тип события | `billing.account.created` |
| `BILLING_URL` | URL сервиса биллинга | `http://billing-service:8080` |

## Запуск

### Docker Compose

```bash
docker compose up -d
```

### Локально

```bash
go run ./main.go
```

## Эндпоинты

- `GET /metrics` — метрики Prometheus (используется и как health-check)
- `/api/v1/profile/...` — операции с профилями

## Связанные репозитории

Инфраструктура всего проекта (k8s, Helm, docker-compose всего стека): https://github.com/n-mark/final-project