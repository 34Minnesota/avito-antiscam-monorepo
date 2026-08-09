# Кейс 5: «Антискам тренажер»

## Запуск

Клонируйте репозиторий и перейдите в него:

```bash
git clone https://github.com/34Minnesota/avito-antiscam-monorepo.git
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
