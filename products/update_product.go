package products

import (
	"encoding/json"
	"io"
	"net/http"

	"shopping_list/db"

	"github.com/gorilla/mux"
)

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data requestData
	err = json.Unmarshal(body, &data) // Parse JSON
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := `UPDATE products SET title = ?, amount_type = ?, price = ? WHERE id = ?`
	_, queryErr := db.DB.Exec(query, data.Title, data.AmountType, data.Price, id)
	if queryErr != nil {
		http.Error(w, "Failed to update the product", http.StatusBadRequest)
		return
	}

	// Create a response struct with data
	response := defaultResponse{
		Data:   "Product successfully updated",
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
