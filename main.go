package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]string{
			"status": "ok",
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("Server running on port :6969")
	http.ListenAndServe(":6969", nil)
}
