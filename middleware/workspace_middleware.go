package middleware

import (
	"net/http"
	"shopping_list/db"
	"strconv"

	"github.com/gorilla/mux"
)

// WorkspaceMiddleware checks if the workspace_id exists in the request
// and verifies that the user has access to this workspace
func WorkspaceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get workspace ID from URL parameters
		vars := mux.Vars(r)
		workspaceID, exists := vars["workspace_id"]
		if !exists {
			http.Error(w, "Workspace ID is required", http.StatusBadRequest)
			return
		}

		// Check if workspace exists
		var workspaceExists int
		err := db.DB.QueryRow("SELECT id FROM workspaces WHERE id = ? AND deleted_at IS NULL", workspaceID).Scan(&workspaceExists)
		if err != nil {
			http.Error(w, "Workspace not found", http.StatusNotFound)
			return
		}

		// Get user ID from the token
		userID, err := ExtractUserIDFromToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user has access to this workspace
		// Either as the owner or as a workspace user
		var hasAccess bool
		err = db.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM workspaces 
				WHERE id = ? AND user_id = ? AND deleted_at IS NULL
				UNION
				SELECT 1 FROM workspace_users 
				WHERE workspace_id = ? AND user_id = ? AND deleted_at IS NULL
			)`, workspaceID, userID, workspaceID, userID).Scan(&hasAccess)

		if err != nil || !hasAccess {
			http.Error(w, "You don't have access to this workspace", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// ProductWorkspaceMiddleware checks if the product belongs to the specified workspace
func ProductWorkspaceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		workspaceID := vars["workspace_id"]
		productID, exists := vars["id"]

		// Only check product workspace if product ID exists in the request
		if exists && productID != "" {
			// Check if product exists and belongs to the specified workspace
			var productWorkspaceID int
			err := db.DB.QueryRow("SELECT workspace_id FROM products WHERE id = ? AND deleted_at IS NULL", productID).Scan(&productWorkspaceID)
			if err != nil {
				http.Error(w, "Product not found", http.StatusNotFound)
				return
			}

			workspaceIDInt, _ := strconv.Atoi(workspaceID)
			if productWorkspaceID != workspaceIDInt {
				http.Error(w, "Product does not belong to this workspace", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}

// CombinedWorkspaceMiddleware combines both workspace and product workspace checks
func CombinedWorkspaceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return TokenAuthMiddleware(WorkspaceMiddleware(ProductWorkspaceMiddleware(next)))
}
