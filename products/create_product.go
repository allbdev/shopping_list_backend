package products

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"shopping_list/db"
	"strconv"
)

type requestData struct {
	Title     string  `json:"title"`
	AmoutType string  `json:"amount_type"`
	Price     float32 `json:"price"`
}

type defaultResponse struct {
	Status string
	Data   string
}

func CreateProdcut(w http.ResponseWriter, r *http.Request) {

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data requestData
	err = json.Unmarshal(body, &data) // Parse JSON
	fmt.Println("Error:", err)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO products (title, amount_type, price) VALUES (?, ?, ?)`
	result, err := db.DB.Exec(query, data.Title, data.AmoutType, data.Price)
	if err != nil {
		http.Error(w, "Failed to insert data", http.StatusBadRequest)
		return
	}

	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	// Create a response struct with data
	response := defaultResponse{
		Data:   "Inserted product with ID: " + strconv.Itoa(int(lastInsertID)) + ", Rows Affected: " + strconv.Itoa(int(rowsAffected)),
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
