package utils

import (
	"fmt"
	"net/http"
	"os"
)


func EnableCORS(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS_ORIGIN comes from the process env (main.go loads .env once at
		// startup locally; Docker injects it directly) — no per-request file
		// reads, and a missing file must never crash the server
		corsOrigin := os.Getenv("CORS_ORIGIN")

		// Set CORS headers for the preflight request and Auth
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Set this if you're using credentials (cookies, HTTP authentication)
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)

		fmt.Printf("CORS set \n\n")

		if r.Method == "OPTIONS" {
			// Preflight request, respond with 200 OK
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the original handler
		next.ServeHTTP(w, r)
	})
}
