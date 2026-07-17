package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) monthlySummaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	query := `
	SELECT
		EXTRACT(YEAR FROM date)::INT AS year,
		EXTRACT(MONTH FROM date)::INT AS month,
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
		COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense,
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)
		- COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS balance,
		COUNT(*) AS count
	FROM transactions
	GROUP BY year, month
	ORDER BY year DESC, month DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}
	defer rows.Close()

	summaries := make([]MonthlySummary, 0)
	for rows.Next() {
		var m MonthlySummary
		err := rows.Scan(
			&m.Year,
			&m.Month,
			&m.TotalIncome,
			&m.TotalExpense,
			&m.Balance,
			&m.Count,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan data")
			return
		}
		summaries = append(summaries, m)
	}

	if rows.Err() != nil {
		writeError(w, http.StatusInternalServerError, "failed to read monthly summary")
		return
	}

	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) categorySummaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	typeFilter := r.URL.Query().Get("type")
	if typeFilter != "" && !validTypes[typeFilter] {
		writeError(w, http.StatusBadRequest, "invalid type")
		return
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

	joinConditions := []string{"t.category_id = c.id"}
	args := make([]any, 0)

	if typeFilter != "" {
		args = append(args, typeFilter)
		joinConditions = append(joinConditions, "t.type = $"+strconv.Itoa(len(args)))
	}
	if startDate != "" {
		args = append(args, startDate)
		joinConditions = append(joinConditions, "t.date >= $"+strconv.Itoa(len(args)))
	}
	if endDate != "" {
		args = append(args, endDate)
		joinConditions = append(joinConditions, "t.date <= $"+strconv.Itoa(len(args)))
	}

	query := `
	SELECT
		c.id AS category_id,
		c.name AS category_name,
		COALESCE(SUM(t.amount), 0) AS total_amount,
		COUNT(t.id) AS count
	FROM categories c
	LEFT JOIN transactions t ON ` + strings.Join(joinConditions, " AND ") + `
	GROUP BY c.id, c.name
	ORDER BY total_amount DESC
	`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}
	defer rows.Close()

	summaries := make([]CategorySummary, 0)
	for rows.Next() {
		var c CategorySummary
		err := rows.Scan(
			&c.CategoryID,
			&c.CategoryName,
			&c.TotalAmount,
			&c.Count,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan data")
			return
		}
		summaries = append(summaries, c)
	}

	if rows.Err() != nil {
		writeError(w, http.StatusInternalServerError, "failed to read category summary")
		return
	}

	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) balanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	query := `
	SELECT
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
		COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense,
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)
		- COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS balance
	FROM transactions
	`

	var b BalanceSummary
	err := s.db.QueryRow(ctx, query).Scan(
		&b.TotalIncome,
		&b.TotalExpense,
		&b.Balance,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch data")
		return
	}

	writeJSON(w, http.StatusOK, b)
}
