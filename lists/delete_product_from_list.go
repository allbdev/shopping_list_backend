package lists

import (
	"encoding/json"
	"net/http"
	"shopping_list/db"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// DeleteProductFromList handles the soft deletion of a product from a list
func DeleteProductFromList(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID, list ID, and product ID from URL parameters
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

	productID, err := strconv.Atoi(vars["product_id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Check if the list exists and belongs to the specified workspace
	var listExists bool
	err = db.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM lists WHERE id = ? AND workspace_id = ? AND deleted_at IS NULL)",
		listID, workspaceID,
	).Scan(&listExists)

	if err != nil || !listExists {
		http.Error(w, "List not found in this workspace", http.StatusNotFound)
		return
	}

	// Soft delete the product from the list by setting the deleted_at timestamp
	_, err = db.DB.Exec(
		"UPDATE list_products SET deleted_at = ? WHERE list_id = ? AND product_id = ? AND deleted_at IS NULL",
		time.Now(), listID, productID,
	)
	if err != nil {
		http.Error(w, "Error deleting product from list: "+err.Error(), http.StatusInternalServerError)
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
