package workspaces

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"shopping_list/db"

	"github.com/gorilla/mux"
)

func RemoveUserFromWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user is part of the workspace
	var existingUserID int
	selectErr := db.DB.QueryRow("SELECT user_id FROM workspace_users WHERE user_id = ? AND workspace_id = ? AND deleted_at IS NULL", userID, workspaceID).Scan(&existingUserID)
	if selectErr != nil {
		http.Error(w, "User is not part of this workspace", http.StatusNotFound)
		return
	}

	// Soft delete the user from the workspace by setting deleted_at
	_, updateErr := db.DB.Exec("UPDATE workspace_users SET deleted_at = ? WHERE user_id = ? AND workspace_id = ?", time.Now(), userID, workspaceID)
	if updateErr != nil {
		http.Error(w, "Error removing user from workspace", http.StatusInternalServerError)
		return
	}

	response := defaultResponse{
		Data:   "User successfully removed from workspace",
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
