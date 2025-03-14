package products

import (
	"encoding/json"
	"io"
	"net/http"
	"shopping_list/db"
)

type requestData struct {
	Title       string  `json:"title"`
	AmountType  string  `json:"amount_type"`
	Price       float32 `json:"price"`
	WorkspaceID int     `json:"workspace_id"`
}

type ProductStruct struct {
	Id          int     `json:"id"`
	Title       string  `json:"title"`
	AmountType  string  `json:"amount_type"`
	Price       float32 `json:"price"`
	WorkspaceID int     `json:"workspace_id"`
}

type defaultResponse struct {
	Status string
	Data   string
}

type productResponse struct {
	Status string
	Data   ProductStruct
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {

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

	// Check if workspace_id exists in the request body
	if data.WorkspaceID == 0 {
		http.Error(w, "Workspace ID is required", http.StatusBadRequest)
		return
	}

	var workspaceExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = ? AND deleted_at IS NULL)", data.WorkspaceID).Scan(&workspaceExists)
	if err != nil {
		http.Error(w, "Failed to verify workspace existence", http.StatusInternalServerError)
		return
	}
	if !workspaceExists {
		http.Error(w, "Workspace does not exist", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO products (title, amount_type, price, workspace_id) VALUES (?, ?, ?, ?)`
	result, queryErr := db.DB.Exec(query, data.Title, data.AmountType, data.Price, data.WorkspaceID)
	if queryErr != nil {
		http.Error(w, "Failed to create the product", http.StatusBadRequest)
		return
	}

	lastInsertID, _ := result.LastInsertId()

	// Create a response struct with data
	response := productResponse{
		Data: ProductStruct{
			Id:          int(lastInsertID),
			Title:       data.Title,
			AmountType:  data.AmountType,
			Price:       data.Price,
			WorkspaceID: data.WorkspaceID,
		},
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
