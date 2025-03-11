package auth

import (
	"encoding/json"
	"net/http"
	"shopping_list/db"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user UserLogin

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var hashedPassword string
	var name string
	err := db.DB.QueryRow("SELECT password, name FROM users WHERE email = ? AND deleted_at IS NULL", user.Email).Scan(&hashedPassword, &name)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	})

	tokenString, err := token.SignedString([]byte("your_secret_key")) // Replace with your secret key
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Store token in the database (you may need to create a new column for tokens)
	_, err = db.DB.Exec("UPDATE users SET token = ? WHERE email = ?", tokenString, user.Email)
	if err != nil {
		http.Error(w, "Error storing token", http.StatusInternalServerError)
		return
	}

	// Respond with the token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "name": name, "email": user.Email})
}
