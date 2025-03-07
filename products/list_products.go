package products

import (
	"encoding/json"
	"net/http"
	"shopping_list/db"
)

type getResponse struct {
	Status string
	Data   []Product
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	var products []Product

	name := "%" + r.URL.Query().Get("name") + "%"

	rows, err := db.DB.Query("SELECT * FROM products WHERE title LIKE ?", name)

	if err != nil {
		http.Error(w, "Failed to query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Title, &product.AmoutType, &product.Price); err != nil {
			http.Error(w, "Failed to create product list", http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to create product list", http.StatusInternalServerError)
		return
	}

	// Create a response struct with data
	response := getResponse{
		Data:   products,
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
