# Mini Notes API

Simple note management API.

## Thing I learned:

- Partial update in Go and Postgres
- `COALESCE` function in sql
- More advance sql query in search note feature: ILIKE, wildcard(%), string concat (||), type casting(::bool), etc.

---

## How to run

1. Setup database

```bash
docker compose up -d
make migrate-up
```

2. Start server

```bash
make run
```

3. Test Api

Open [HTML documentation](docs/mini-notes-documentation.html) in your browser to try the endpoints.
