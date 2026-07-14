package main

import "time"

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}
