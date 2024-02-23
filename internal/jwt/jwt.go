package jwt

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getSecretKey() []byte {
	secretKeyName := "JWT_SECRET"
	if val, ok := os.LookupEnv("JWT_SECRET"); ok {
		secretKeyName = val
	}

	return []byte(secretKeyName)
}

func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
			"exp": time.Now().Add(time.Minute * 15).Unix(),
		})

	tokenString, err := token.SignedString(getSecretKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func AuthenticateUser(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	tokenString := authHeader[len("Bearer "):]
	token, err := VerifyToken(tokenString)

	if err != nil {
		return "", err
	}

	username, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return username, nil
}
