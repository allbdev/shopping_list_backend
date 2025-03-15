package workspaces

import (
	"encoding/json"
	"net/http"

	"shopping_list/db"
	"shopping_list/middleware"

	"github.com/gorilla/mux"
)

func GetWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the token
	userID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]

	var workspace Workspace
	err = db.DB.QueryRow("SELECT id, name, created_at, updated_at, deleted_at, user_id FROM workspaces WHERE id = ? AND user_id = ? AND deleted_at IS NULL", workspaceID, userID).Scan(&workspace.ID, &workspace.Name, &workspace.CreatedAt, &workspace.UpdatedAt, &workspace.DeletedAt, &workspace.UserID)
	if err != nil {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	response := defaultResponse{
		Data:   workspace,
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
