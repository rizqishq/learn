# Personal Expense Tracker API

Simple personal income/expense tracking API.

## Thing I learned:

- PostgreSQL `DATE` type and scanning it in Go
- Dynamic `WHERE` for query parameter filters
- SQL aggregation for summary endpoints:
  - `EXTRACT(YEAR/MONTH FROM date)` — pull year/month pieces out of a `DATE` (result is float, cast with `::INT`)
  - `CASE WHEN ... THEN ... ELSE ... END` — per-row if/else so one `SUM` can total only income or only expense
  - `SUM(...)` — one total per group; returns `NULL` when the group has **zero rows** (not when values are 0)
  - `COALESCE(x, 0)` — turn that `NULL` into `0` so JSON never leaks null totals
  - `GROUP BY` — bucket rows by a key, emit one summary row per bucket (key can be an expression/alias, not only a physical column)
  - `LEFT JOIN` + filters on `ON` — keep every category even with zero transactions; put type/date filters on `ON` so empty categories are not dropped
  - `COUNT(t.id)` vs `COUNT(*)` — count only non-NULL join matches so zero-transaction categories stay at `0`

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
