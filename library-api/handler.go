package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createAuthorHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var req CreateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name cant be empty")
		return
	}

	query := `
	INSERT INTO authors (name)
	VALUES ($1)
	RETURNING id, name, created_at
	`

	var author Authors
	err := s.db.QueryRow(ctx, query, req.Name).Scan(
		&author.ID,
		&author.Name,
		&author.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "author already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create author")
		return
	}

	writeJSON(w, http.StatusCreated, author)
}
