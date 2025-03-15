package workspaces

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shopping_list/db"
	"shopping_list/middleware"

	"github.com/gorilla/mux"
)

func AddUserToWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the token
	loggedInUserID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	var userExists int
	err = db.DB.QueryRow("SELECT id FROM users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&userExists)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user ID is the same as the logged-in user ID
	if userID == loggedInUserID {
		http.Error(w, "Cannot add yourself to the workspace", http.StatusForbidden)
		return
	}

	// Check if the user is already part of the workspace
	var existingUserID int
	err = db.DB.QueryRow("SELECT user_id FROM workspace_users WHERE user_id = ? AND workspace_id = ? AND deleted_at IS NULL", userID, workspaceID).Scan(&existingUserID)
	if err == nil {
		http.Error(w, "User is already part of this workspace", http.StatusConflict)
		return
	}

	// Add the user to the workspace
	_, err = db.DB.Exec("INSERT INTO workspace_users (user_id, workspace_id) VALUES (?, ?)", userID, workspaceID)
	if err != nil {
		http.Error(w, "Error adding user to workspace", http.StatusInternalServerError)
		return
	}

	response := defaultResponse{
		Data:   "User successfully added to workspace",
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
