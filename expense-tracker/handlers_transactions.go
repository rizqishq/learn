package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const dateLayout = "2006-01-02"

var validTypes = map[string]bool{
	"income":  true,
	"expense": true,
}

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
	if !validTypes[req.Type] {
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

	typeFilter := r.URL.Query().Get("type")
	if typeFilter != "" && !validTypes[typeFilter] {
		writeError(w, http.StatusBadRequest, "invalid type")
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID int64
	if categoryIDStr != "" {
		id, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category id")
			return
		}
		categoryID = id
	}

	startDate := r.URL.Query().Get("start_date")
	if startDate != "" {
		if _, err := time.Parse(dateLayout, startDate); err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	endDate := r.URL.Query().Get("end_date")
	if endDate != "" {
		if _, err := time.Parse(dateLayout, endDate); err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	q := r.URL.Query().Get("q")

	conditions := make([]string, 0)
	args := make([]any, 0)

	if typeFilter != "" {
		args = append(args, typeFilter)
		conditions = append(conditions, "t.type = $"+strconv.Itoa(len(args)))
	}

	if categoryID != 0 {
		args = append(args, categoryID)
		conditions = append(conditions, "t.category_id = $"+strconv.Itoa(len(args)))
	}

	if startDate != "" {
		args = append(args, startDate)
		conditions = append(conditions, "t.date >= $"+strconv.Itoa(len(args)))
	}

	if endDate != "" {
		args = append(args, endDate)
		conditions = append(conditions, "t.date <= $"+strconv.Itoa(len(args)))
	}

	if q != "" {
		args = append(args, q)
		conditions = append(conditions, "t.note ILIKE '%' || $"+strconv.Itoa(len(args))+" || '%'")
	}

	query := `
	SELECT
		t.id, t.type, t.amount, t.note, t.date, t.created_at, t.updated_at,
		c.id, c.name
	FROM transactions t
	JOIN categories c ON c.id = t.category_id
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.date DESC, t.id DESC"

	rows, err := s.db.Query(ctx, query, args...)
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

func (s *Server) getTransactionByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	query := `
	SELECT
		t.id, t.type, t.amount, t.note, t.date, t.created_at, t.updated_at,
		c.id, c.name
	FROM transactions t
	JOIN categories c ON c.id = t.category_id
	WHERE t.id = $1
	`

	var t Transaction
	var date time.Time
	err = s.db.QueryRow(ctx, query, id).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}

	t.Date = date.Format(dateLayout)

	writeJSON(w, http.StatusOK, t)
}

func (s *Server) updateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateTransactionRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.CategoryID == nil && req.Amount == nil && req.Date == nil && req.Note == nil && req.Type == nil {
		writeError(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	if req.Type != nil && !validTypes[*req.Type] {
		writeError(w, http.StatusBadRequest, "type must be income or expense")
		return
	}
	if req.Amount != nil && *req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}
	if req.Date != nil {
		if _, err := time.Parse(dateLayout, *req.Date); err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	query := `
	WITH updated_tx AS (
		UPDATE transactions
		SET
			category_id = COALESCE($1, category_id),
			type = COALESCE($2, type),
			amount = COALESCE($3, amount),
			note = COALESCE($4, note),
			date = COALESCE($5, date),
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, category_id, type, amount, note, date, created_at, updated_at
	)
	SELECT
		t.id, t.type, t.amount, t.note, t.date, t.created_at, t.updated_at,
		c.id, c.name
	FROM updated_tx t
	JOIN categories c ON c.id = t.category_id
	`

	var t Transaction
	var date time.Time
	err = s.db.QueryRow(ctx, query, req.CategoryID, req.Type, req.Amount, req.Note, req.Date, id).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusNotFound, "category does not exist")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update transaction")
		return
	}

	t.Date = date.Format(dateLayout)
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) deleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	query := `
	DELETE FROM transactions
	WHERE id = $1
	`

	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
