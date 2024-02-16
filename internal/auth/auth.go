package auth

import (
	"fmt"
	"net/http"
	"os"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
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
