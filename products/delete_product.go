package products

import (
	"encoding/json"
	"net/http"
	"time"

	"shopping_list/db"

	"github.com/gorilla/mux"
)

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	query := `UPDATE products SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := db.DB.Exec(query, time.Now(), id)
	if err != nil {
		http.Error(w, "Failed to delete the product", http.StatusBadRequest)
		return
	}

	// Create a response struct with data
	response := defaultResponse{
		Data:   "Product successfully deleted",
		Status: "Success",
	}

	// Set the response header to indicate the content is JSON
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code (optional)
	w.WriteHeader(http.StatusOK)

	// Encode the struct into JSON and write it to the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If encoding fails, return an error message
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
