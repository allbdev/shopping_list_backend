package lists

import (
	"encoding/json"
	"net/http"
	"shopping_list/db"
	"shopping_list/middleware"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ListStatus represents the possible status values for a list
const (
	ListStatusDeleted   = 0
	ListStatusActive    = 1
	ListStatusCompleted = 2
)

// UpdateListStatusRequest represents the request body for updating a list's status
type UpdateListStatusRequest struct {
	Status int `json:"status"`
}

// UpdateListStatus handles updating the status of a list
func UpdateListStatus(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID and list ID from URL parameters
	vars := mux.Vars(r)
	workspaceID, err := strconv.Atoi(vars["workspace_id"])
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(vars["list_id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Get user ID from the token
	userID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req UpdateListStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status value
	if req.Status < ListStatusDeleted || req.Status > ListStatusCompleted {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Check if the list exists and belongs to the specified workspace
	var listExists bool
	var listOwnerID int
	err = db.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM lists WHERE id = ? AND workspace_id = ? AND deleted_at IS NULL), user_id FROM lists WHERE id = ?",
		listID, workspaceID, listID,
	).Scan(&listExists, &listOwnerID)

	if err != nil || !listExists {
		http.Error(w, "List not found in this workspace", http.StatusNotFound)
		return
	}

	// Check if the user has access to modify this list
	// Either they are the list owner or they have access to the workspace
	var hasAccess bool
	if listOwnerID == userID {
		hasAccess = true
	} else {
		err = db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM workspaces 
				WHERE id = ? AND user_id = ? AND deleted_at IS NULL
				UNION
				SELECT 1 FROM workspace_users 
				WHERE workspace_id = ? AND user_id = ? AND deleted_at IS NULL
			)`, workspaceID, userID, workspaceID, userID).Scan(&hasAccess)

		if err != nil {
			http.Error(w, "Error checking access", http.StatusInternalServerError)
			return
		}
	}

	if !hasAccess {
		http.Error(w, "You don't have permission to modify this list", http.StatusForbidden)
		return
	}

	// Update the list status
	_, err = db.DB.Exec(
		"UPDATE lists SET status = ?, updated_at = ? WHERE id = ?",
		req.Status, time.Now(), listID,
	)
	if err != nil {
		http.Error(w, "Error updating list status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If status is "deleted", soft delete the list
	if req.Status == ListStatusDeleted {
		_, err = db.DB.Exec(
			"UPDATE lists SET deleted_at = ? WHERE id = ?",
			time.Now(), listID,
		)
		if err != nil {
			http.Error(w, "Error deleting list: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Retrieve the updated list
	var list ProductList
	err = db.DB.QueryRow(
		"SELECT id, workspace_id, user_id, title, created_at, updated_at, deleted_at FROM lists WHERE id = ?",
		listID,
	).Scan(
		&list.ID, &list.WorkspaceID, &list.UserID, &list.Title, &list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
	)
	if err != nil {
		http.Error(w, "Error retrieving updated list", http.StatusInternalServerError)
		return
	}

	// Return the updated list
	response := struct {
		Status string      `json:"status"`
		Data   ProductList `json:"data"`
	}{
		Status: "Success",
		Data:   list,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
