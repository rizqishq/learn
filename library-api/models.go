package main

import "time"

type Author struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAuthorRequest struct {
	Name string `json:"name"`
}
