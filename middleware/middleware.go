package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// TokenAuthMiddleware checks if the user is logged in based on the JWT token
func TokenAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		tokenString := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		if tokenString == "" {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNoLocation
			}
			return []byte("your_secret_key"), nil // Replace with your secret key
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
