# Кейс 5: «Антискам тренажер»

## Быстрый старт через Make

Клонируйте репозиторий и перейдите в него:

```bash
git clone git@github.com:34Minnesota/avito-antiscam-monorepo.git
cd avito-antiscam-monorepo
```

Создайте локальный файл конфигурации:

```bash
cp .env.example .env
```

Запустите базу данных, примените миграции и поднимите API:

```bash
make db-up
make migrate-up
make run
```

После запуска:

- API: [http://localhost:8080](http://localhost:8080);
- health-check: [http://localhost:8080/healthz](http://localhost:8080/healthz);
- Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

## Запуск вручную

Если `make` недоступен, тот же сценарий можно выполнить напрямую.

### 1. Клонирование и конфигурация

```bash
git clone git@github.com:34Minnesota/avito-antiscam-monorepo.git
cd avito-antiscam-monorepo

cp .env.example .env
set -a
source .env
set +a

export PROJECT_ROOT="$PWD"
```

### 2. Запуск PostgreSQL

```bash
docker compose up -d postgres
```

Проверить состояние контейнеров:

```bash
docker compose ps
```

### 3. Применение миграций

```bash
docker compose run --rm postgres-migrate \
  -path /migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" \
  up
```

### 4. Запуск API

```bash
cd backend
go run ./cmd/antiscam
```

## Полезные команды Make

| Команда | Назначение |
| --- | --- |
| `make db-up` | Запустить PostgreSQL |
| `make db-down` | Остановить PostgreSQL |
| `make ps` | Показать состояние контейнеров |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откатить миграции |
| `make migrate-create seq=init` | Создать новую миграцию |
| `make run` | Запустить API |
| `make test` | Запустить тесты с покрытием |
| `make lint` | Запустить линтер |
| `make swag` | Перегенерировать Swagger-документацию |

## Остановка

Остановить API можно сочетанием `Ctrl+C`. PostgreSQL останавливается отдельно:

```bash
make db-down
```

При ручном запуске:

```bash
docker compose down postgres
```

Данные базы сохраняются в Docker volume `postgres-data`, поэтому остановка
контейнера не удаляет данные из volume.
