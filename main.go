package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("Failed to create connection pool: ", err)
	}
	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Database connected succesfully")

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			resp := map[string]string{
				"status":   "error",
				"database": "disconnected",
			}
			writeJSON(w, http.StatusServiceUnavailable, resp)
		}
		resp := map[string]string{
			"status":   "ok",
			"database": "connected",
		}
		writeJSON(w, http.StatusOK, resp)
	})

	log.Println("Server running on port :6969")
	http.ListenAndServe(":6969", nil)
}
