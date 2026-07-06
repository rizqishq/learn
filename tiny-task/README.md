# Tiny Tasks API

A simple task management API built with Go.

## Things I learned while building this project

- REST API design in Go
- PostgreSQL with database/sql + pgx
- Raw SQL queries and migrations
- URL path parameter parsing
- JSON encoding/decoding

---

## How to Run

1. Set up the database

```bash
docker compose up -d
make migrate-up
```

2. Start the server

```bash
make run
```

3. Test the API

Open the [HTML documentation](docs/tiny-tasks-documentation.html) in your browser to try the endpoints.
