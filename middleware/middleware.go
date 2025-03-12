package middleware

import (
	"net/http"
	"shopping_list/db"
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

// ExtractUserIDFromToken extracts the user ID from the JWT token
func ExtractUserIDFromToken(r *http.Request) (int, error) {
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	if tokenString == "" {
		return 0, http.ErrNoLocation
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your_secret_key"), nil // Replace with your secret key
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims["email"].(string); ok {
			user, err := db.DB.Query("SELECT id FROM users WHERE email = ?", email)
			if err != nil {
				return 0, err
			}
			defer user.Close()

			if user.Next() {
				var id int
				if err := user.Scan(&id); err != nil {
					return 0, err
				}
				return id, nil
			}
		}
	}

	return 0, err
}
