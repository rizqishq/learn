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

type UpdateTransactionRequest struct {
	CategoryID *int64  `json:"category_id"`
	Type       *string `json:"type"`
	Amount     *int64  `json:"amount"`
	Note       *string `json:"note"`
	Date       *string `json:"date"`
}

type MonthlySummary struct {
	Year         int   `json:"year"`
	Month        int   `json:"month"`
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
	Balance      int64 `json:"balance"`
	Count        int   `json:"count"`
}

type CategorySummary struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	TotalAmount  int64  `json:"total_amount"`
	Count        int    `json:"count"`
}

type BalanceSummary struct {
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
	Balance      int64 `json:"balance"`
}
