package utils

import (
	"net/http"
	"net/url"
	"os"
)

// isLoopbackOrigin reports whether an Origin header points at the local machine
// (http://localhost:5173, http://127.0.0.1:8016, http://[::1]:5173, ...)
func isLoopbackOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := parsed.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func EnableCORS(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS_ORIGIN comes from the process env (main.go loads .env once at
		// startup locally; Docker injects it directly) — no per-request file
		// reads, and a missing file must never crash the server
		corsOrigin := os.Getenv("CORS_ORIGIN")

		// Local dev (ENV != PROD): the page can sit on a different port than the
		// API (vite on 5173 calling Go on 8016) — echo loopback origins so those
		// calls aren't blocked. Prod keeps the single configured origin.
		if os.Getenv("ENV") != "PROD" {
			if origin := r.Header.Get("Origin"); origin != "" && isLoopbackOrigin(origin) {
				corsOrigin = origin
			}
		}

		// Set CORS headers for the preflight request and Auth
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Set this if you're using credentials (cookies, HTTP authentication)
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)

		// browsers cache a preflight per URL — vary on Origin so a loopback echo
		// never poisons the cache for the prod origin (and vice versa)
		w.Header().Add("Vary", "Origin")

		if r.Method == "OPTIONS" {
			// Preflight request, respond with 200 OK
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the original handler
		next.ServeHTTP(w, r)
	})
}
