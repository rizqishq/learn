package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const dateLayout = "2006-01-02"

func (s *Server) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var req CreateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type cannot be empty")
		return
	}
	if req.Type != "income" && req.Type != "expense" {
		writeError(w, http.StatusBadRequest, "type must be income or expense")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}
	if req.Date == "" {
		req.Date = time.Now().Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	query := `
	WITH new_tx AS (
		INSERT INTO transactions (category_id, type, amount, note, date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, category_id, type, amount, note, date, created_at, updated_at
	)
	SELECT
		t.id, t.type, t.amount, t.note, t.date, t.created_at, t.updated_at,
		c.id, c.name
	FROM new_tx t
	JOIN categories c ON c.id = t.category_id
	`

	var t Transaction
	var date time.Time
	err = s.db.QueryRow(ctx, query, req.CategoryID, req.Type, req.Amount, req.Note, req.Date).Scan(
		&t.ID,
		&t.Type,
		&t.Amount,
		&t.Note,
		&date,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Category.ID,
		&t.Category.Name,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusNotFound, "category does not exist")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}

	t.Date = date.Format(dateLayout)
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	query := `
	SELECT
		t.id, t.type, t.amount, t.note, t.date, t.created_at, t.updated_at,
		c.id, c.name
	FROM transactions t
	JOIN categories c ON c.id = t.category_id
	ORDER BY t.id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	var date time.Time
	for rows.Next() {
		var t Transaction
		err := rows.Scan(
			&t.ID,
			&t.Type,
			&t.Amount,
			&t.Note,
			&date,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.Category.ID,
			&t.Category.Name,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch data")
			return
		}
		t.Date = date.Format(dateLayout)
		transactions = append(transactions, t)
	}

	writeJSON(w, http.StatusOK, transactions)
}
