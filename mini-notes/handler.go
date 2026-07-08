package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title can not be empty")
		return
	}

	query := `
	INSERT INTO notes (title, body)
	VALUES ($1, $2)
	RETURNING id, title, body, archived, created_at, updated_at
	`

	var note Note
	err := s.db.QueryRow(ctx, query, req.Title, req.Body).Scan(
		&note.ID,
		&note.Title,
		&note.Body,
		&note.Archived,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

func (s *Server) getNoteByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	query := `
	SELECT id, title, body, archived, created_at, updated_at
	FROM notes
	WHERE id = $1
	`

	var note Note
	err = s.db.QueryRow(ctx, query, id).Scan(
		&note.ID,
		&note.Title,
		&note.Body,
		&note.Archived,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "operation failed")
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (s *Server) updateNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateNoteRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Body == nil && req.Archived == nil {
		writeError(w, http.StatusBadRequest, "no field to update")
		return
	}

	if req.Title != nil && *req.Title == "" {
		writeError(w, http.StatusBadRequest, "title can not be empty")
		return
	}
	if req.Body != nil && *req.Body == "" {
		writeError(w, http.StatusBadRequest, "body can not be empty")
		return
	}

	query := `
	UPDATE notes
	SET title = COALESCE($1, title), body = COALESCE($2, body), archived = COALESCE($3, archived), updated_at = NOW()
	WHERE id = $4
	RETURNING id, title, body, archived, created_at, updated_at
	`

	var note Note
	err = s.db.QueryRow(ctx, query, req.Title, req.Body, req.Archived, id).Scan(
		&note.ID,
		&note.Title,
		&note.Body,
		&note.Archived,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (s *Server) deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	query := `
	DELETE FROM notes
	WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) searchNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	q := r.URL.Query().Get("q")
	archivedParam := r.URL.Query().Get("archived")

	var archived *bool
	if archivedParam != "" {
		parsed, err := strconv.ParseBool(archivedParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid archived value")
			return
		}
		archived = &parsed
	}

	query := `
	SELECT id, title, body, archived, created_at, updated_at
	FROM notes
	WHERE ($1 = '' OR title ILIKE '%'||$1||'%' OR body ILIKE '%'||$1||'%')
		AND ($2::bool IS NULL OR archived = $2)
	ORDER BY updated_at DESC
	`

	rows, err := s.db.Query(ctx, query, q, archived)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var note Note
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Body,
			&note.Archived,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read note")
			return
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read note")
		return
	}

	writeJSON(w, http.StatusOK, notes)
}
