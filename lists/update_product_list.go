package lists

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shopping_list/db"
	"shopping_list/middleware"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// UpdateProductListRequest represents the request body for updating a product list
type UpdateProductListRequest struct {
	Title    string               `json:"title"`
	Status   int                  `json:"status"`
	Products []ListProductRequest `json:"products"`
}

// UpdateProductList handles updating the title, status, and products of a list
func UpdateProductList(w http.ResponseWriter, r *http.Request) {
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
	var req UpdateProductListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status value
	if req.Status < ListStatusDeleted || req.Status > ListStatusCompleted {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Rollback if not committed

	// Update the list title and status
	_, err = tx.Exec(
		"UPDATE lists SET title = ?, status = ?, updated_at = ? WHERE id = ? AND workspace_id = ? AND user_id = ? AND deleted_at IS NULL",
		req.Title, req.Status, time.Now(), listID, workspaceID, userID,
	)
	if err != nil {
		http.Error(w, "Error updating list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle products in the list
	existingProducts := ""
	for index, product := range req.Products {
		existingProducts += strconv.Itoa(product.ProductID)
		if index < len(req.Products)-1 {
			existingProducts += ", "
		}

		// Check if the product already exists in the list
		var exists bool
		err := tx.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM list_products WHERE list_id = ? AND product_id = ? AND deleted_at IS NULL)",
			listID, product.ProductID,
		).Scan(&exists)

		if err != nil {
			http.Error(w, "Error checking product existence: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if exists {
			// Update existing product
			_, err = tx.Exec(
				"UPDATE list_products SET quantity = ?, updated_at = ? WHERE list_id = ? AND product_id = ?",
				product.Quantity, time.Now(), listID, product.ProductID,
			)
		} else {
			// Insert new product
			_, err = tx.Exec(
				"INSERT INTO list_products (list_id, product_id, quantity, checked, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
				listID, product.ProductID, product.Quantity, false, time.Now(), time.Now(),
			)
		}

		if err != nil {
			http.Error(w, "Error updating product in list: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Soft delete products that are not in the request
	if len(req.Products) > 0 {
		// Build a proper query with placeholders for each product ID
		query := "UPDATE list_products SET deleted_at = ? WHERE list_id = ? AND deleted_at IS NULL"

		// Only add the NOT IN clause if there are products to exclude
		placeholders := make([]interface{}, 0, len(req.Products)+2)
		placeholders = append(placeholders, time.Now(), listID)

		if len(req.Products) > 0 {
			query += " AND product_id NOT IN ("
			for i, product := range req.Products {
				if i > 0 {
					query += ", "
				}
				query += "?"
				placeholders = append(placeholders, product.ProductID)
			}
			query += ")"
		}

		_, err = tx.Exec(query, placeholders...)
	} else {
		// If no products in request, mark all products as deleted
		_, err = tx.Exec(
			"UPDATE list_products SET deleted_at = ? WHERE list_id = ? AND deleted_at IS NULL",
			time.Now(), listID,
		)
	}
	if err != nil {
		fmt.Print(existingProducts)
		http.Error(w, "Error deleting products from list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := struct {
		Status string `json:"status"`
	}{
		Status: "Success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
