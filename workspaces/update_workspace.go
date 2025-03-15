package workspaces

import (
	"encoding/json"
	"net/http"
	"time"

	"shopping_list/db"
	"shopping_list/middleware"

	"github.com/gorilla/mux"
)

func UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
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

	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]

	// Update the workspace in the database
	_, err = db.DB.Exec("UPDATE workspaces SET name = ?, updated_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL", workspace.Name, time.Now(), workspaceID, userID)
	if err != nil {
		http.Error(w, "Error updating workspace", http.StatusInternalServerError)
		return
	}

	response := defaultResponse{
		Data:   workspace,
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
