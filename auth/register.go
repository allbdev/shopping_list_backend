package auth

import (
	"encoding/json"
	"net/http"

	"shopping_list/db"

	"golang.org/x/crypto/bcrypt"
)

type UserRegister struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user UserRegister
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (email, password, name) VALUES (?, ?, ?)", user.Email, hashedPassword, user.Name)
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("User registered successfully")
}
