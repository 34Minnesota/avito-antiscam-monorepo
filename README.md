# Кейс 5: «Антискам тренажер»


## Run

```bash
make db-up
make migrate-up
make run
```

## Run tests

```bash
make test
```

## Run linter

```bash
make lint
```

## Update swagger API documentation

```bash
make swag
```



Frontend:
http://localhost:3000

Backend:
http://localhost:8080

Swagger:
http://localhost:8080/swagger/index.html

Health:
http://localhost:3000/healthz
http://localhost:8080/healthz