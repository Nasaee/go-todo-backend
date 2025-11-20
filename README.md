# Ecommerce API

## Running the API

```bash
docker compose up -d
```

## Prerequisites

- Docker
- [Goose](https://github.com/pressly/goose)

## Running the API

```bash
go run cmd/*.go
```

## Adding a new table

1. Create a new migration file under `internal/db/postgres/migrations` folder or anywhere else

```.env
GOOSE_DBSTRING="host=localhost port=5433 user=postgres password=P@ssw0rd dbname=test sslmode=disable"
GOOSE_DRIVER=postgres
GOOSE_MIGRATION_DIR=./internal/db/postgres/migrations
```

```bash
goose -s create create_products sql
```

2.Run the migrations

```bash
goose up
```
