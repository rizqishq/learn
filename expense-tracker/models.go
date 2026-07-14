package main

import "time"

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TransactionCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Transaction struct {
	ID        int64               `json:"id"`
	Type      string              `json:"type"`
	Amount    int64               `json:"amount"`
	Note      string              `json:"note"`
	Date      string              `json:"date"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Category  TransactionCategory `json:"category"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type CreateTransactionRequest struct {
	CategoryID int64  `json:"category_id"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
	Note       string `json:"note"`
	Date       string `json:"date"`
}
