package workspaces

import (
	"encoding/json"
	"net/http"
	"time"

	"shopping_list/db"
	"shopping_list/middleware"

	"github.com/gorilla/mux"
)

func DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]

	// Get user ID from the token
	userID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Soft delete the workspace by setting deleted_at
	_, err = db.DB.Exec("UPDATE workspaces SET deleted_at = ? WHERE id = ? AND user_id = ?", time.Now(), workspaceID, userID)
	if err != nil {
		http.Error(w, "Error deleting workspace", http.StatusInternalServerError)
		return
	}

	// Create a response struct with data
	response := defaultResponse{
		Data:   "Workspace successfully deleted",
		Status: "Success",
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
