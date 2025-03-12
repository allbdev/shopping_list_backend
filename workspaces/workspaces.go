package workspaces

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"shopping_list/db"
	"shopping_list/middleware" // Import middleware for token validation

	"github.com/gorilla/mux"
)

type Workspace struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	UserID    int        `json:"user_id"` // New field for user ID
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type defaultResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

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

func ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the token
	userID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query("SELECT id, name, created_at, updated_at, deleted_at, user_id FROM workspaces WHERE user_id = ? AND deleted_at IS NULL", userID)
	if err != nil {
		http.Error(w, "Error fetching workspaces", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var workspace Workspace
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.CreatedAt, &workspace.UpdatedAt, &workspace.DeletedAt, &workspace.UserID); err != nil {
			http.Error(w, "Error scanning workspace", http.StatusInternalServerError)
			return
		}
		workspaces = append(workspaces, workspace)
	}

	response := defaultResponse{
		Data:   workspaces,
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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
	_, err = db.DB.Exec("UPDATE workspaces SET name = ?, updated_at = ? WHERE id = ? AND user_id = ?", workspace.Name, time.Now(), workspaceID, userID)
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
