package workspaces

import (
	"encoding/json"
	"net/http"
	"time"

	"shopping_list/db"
	"shopping_list/middleware"
)

func CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var workspace Workspace
	if err := json.NewDecoder(r.Body).Decode(&workspace); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Get user ID from the token
	userID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Insert the new workspace into the database
	result, err := db.DB.Exec("INSERT INTO workspaces (name, user_id, created_at, updated_at) VALUES (?, ?, ?, ?)", workspace.Name, userID, time.Now(), time.Now())
	if err != nil {
		http.Error(w, "Error creating workspace", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving workspace ID", http.StatusInternalServerError)
		return
	}

	workspace.ID = int(id)
	workspace.UserID = userID
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()

	response := defaultResponse{
		Data:   workspace,
		Status: "Success",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
