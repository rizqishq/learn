package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var req CreateCategoryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	query := `
	INSERT INTO categories (name)
	VALUES ($1)
	RETURNING id, name, created_at
	`

	var c Category
	err = s.db.QueryRow(ctx, query, req.Name).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "category already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) listCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	query := `
	SELECT id, name, created_at
	FROM categories
	ORDER BY name
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.CreatedAt,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan data")
			return
		}
		categories = append(categories, c)
	}

	if rows.Err() != nil {
		writeError(w, http.StatusInternalServerError, "failed to read categories")
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	query := `
	DELETE FROM categories
	WHERE id = $1
	`

	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
