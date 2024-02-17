package middleware

import (
	"fmt"
	"net/http"
	"os"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	secretToken := "TOKEN"
	if val, ok := os.LookupEnv("TOKEN"); ok {
		secretToken = val
	}

	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != fmt.Sprintf("Bearer %s", secretToken) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// If it's a preflight request, respond immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Otherwise, pass the request on to the next handler
		next.ServeHTTP(w, r)
	})
}
