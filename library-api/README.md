# Library Management API

Simple Library Management API.

## Thing I learned:

- Working with 2 table
- SQL JOIN query: what is JOIN, and how to JOIN 2 table
- [CTE (Common Table Expression)](https://www.postgresql.org/docs/current/queries-with.html)
- [Postgres errcodes](https://www.postgresql.org/docs/8.2/errcodes-appendix.html)

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

