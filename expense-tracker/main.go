package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	db *pgxpool.Pool
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	log.Println("Database connected")

	srv := Server{db: db}

	http.HandleFunc("GET /health", srv.healthHandler)
	http.HandleFunc("POST /categories", srv.createCategoryHandler)
	http.HandleFunc("GET /categories", srv.listCategoriesHandler)
	http.HandleFunc("DELETE /categories/{id}", srv.deleteCategoryHandler)

	http.HandleFunc("POST /transactions", srv.createTransactionHandler)
	http.HandleFunc("GET /transactions", srv.listTransactionsHandler)
	http.HandleFunc("GET /transactions/{id}", srv.getTransactionByIDHandler)
	http.HandleFunc("PATCH /transactions/{id}", srv.updateTransactionHandler)
	http.HandleFunc("DELETE /transactions/{id}", srv.deleteTransactionHandler)

	http.HandleFunc("GET /summary/monthly", srv.monthlySummaryHandler)
	http.HandleFunc("GET /summary/categories", srv.categorySummaryHandler)
	http.HandleFunc("GET /summary/balance", srv.balanceHandler)

	log.Println("Server running on port :6767")
	log.Fatal(http.ListenAndServe(":6767", nil))
}
