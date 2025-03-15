package workspaces

import (
	"encoding/json"
	"net/http"
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

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ListUsersInWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the token
	loggedInUserID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID := vars["workspace_id"]

	// Check if the logged-in user is the owner of the workspace
	var ownerID int
	err = db.DB.QueryRow("SELECT user_id FROM workspaces WHERE id = ? AND deleted_at IS NULL", workspaceID).Scan(&ownerID)
	if err != nil || ownerID != loggedInUserID {
		http.Error(w, "Unauthorized access to this workspace", http.StatusUnauthorized)
		return
	}

	// Retrieve users in the workspace with user info
	rows, err := db.DB.Query(`
		SELECT u.id, u.name, u.email 
		FROM workspace_users wu
		JOIN users u ON wu.user_id = u.id 
		WHERE wu.workspace_id = ? AND wu.deleted_at IS NULL`, workspaceID)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	response := defaultResponse{
		Data:   users,
		Status: "Success",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
