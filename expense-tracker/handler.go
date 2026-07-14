package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) heatlhHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"success": "ok",
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

	query := `
	INSERT INTO categories (name)
	VALUES ($1)
	RETURNING id, name, created_at
	`

	var c Category
	err = s.db.QueryRow(ctx, query, req).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	writeJSON(w, http.StatusOK, c)
}
