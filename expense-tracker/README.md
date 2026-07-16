# Personal Expense Tracker API

Simple personal income/expense tracking API.

## Thing I learned:

- PostgreSQL `DATE` type and scanning it in Go
- Dynamic `WHERE` for query parameter filters

---

## How to run

1. Setup database

```bash
cp .env.example .env

docker compose up -d
make migrate-up
```

2. Start server

```bash
make run
```
