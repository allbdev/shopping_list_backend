package lists

import (
	"encoding/json"
	"net/http"
	"shopping_list/db"
	"shopping_list/middleware"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ProductList represents a list of products
type ProductList struct {
	ID          int           `json:"id"`
	WorkspaceID int           `json:"workspace_id"`
	UserID      int           `json:"user_id"`
	Title       string        `json:"title"`
	Products    []ListProduct `json:"products,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   *time.Time    `json:"deleted_at,omitempty"`
}

type ListProduct struct {
	ListID    int        `json:"list_id"`
	ProductID int        `json:"product_id"`
	Quantity  int        `json:"quantity"`
	Checked   bool       `json:"checked"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ListProduct represents a product in a list with its quantity
type ListProductRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CreateProductListRequest represents the request body for creating a product list
type CreateProductListRequest struct {
	Title    string               `json:"title"`
	Products []ListProductRequest `json:"products"`
}

// CreateProductList handles the creation of a new product list
func CreateProductList(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID, err := strconv.Atoi(vars["workspace_id"])
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Get user ID from the token
	loggedInUserID, err := middleware.ExtractUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreateProductListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Rollback if not committed

	// Create the product list
	result, err := tx.Exec(
		"INSERT INTO lists (workspace_id, user_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		workspaceID, loggedInUserID, req.Title, time.Now(), time.Now(),
	)
	if err != nil {
		http.Error(w, "Error creating product list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly created product list
	listID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving product list ID", http.StatusInternalServerError)
		return
	}

	// Insert products into the list if provided
	if len(req.Products) > 0 {
		// Verify all products belong to the workspace
		for _, product := range req.Products {
			var productWorkspaceID int
			err := tx.QueryRow("SELECT workspace_id FROM products WHERE id = ? AND deleted_at IS NULL", product.ProductID).Scan(&productWorkspaceID)
			if err != nil {
				http.Error(w, "Product not found: "+strconv.Itoa(product.ProductID), http.StatusBadRequest)
				return
			}
			if productWorkspaceID != workspaceID {
				http.Error(w, "Product does not belong to this workspace: "+strconv.Itoa(product.ProductID), http.StatusBadRequest)
				return
			}

			// Set default quantity if not provided
			quantity := product.Quantity
			if quantity <= 0 {
				quantity = 1
			}

			// Insert product into list
			_, err = tx.Exec(
				"INSERT INTO list_products (list_id, product_id, quantity, checked, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
				listID, product.ProductID, quantity, false, time.Now(), time.Now(),
			)
			if err != nil {
				http.Error(w, "Error adding product to list: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	// Retrieve all products in the list
	var products []ListProduct
	rows, err := db.DB.Query(
		"SELECT product_id, quantity, checked FROM list_products WHERE list_id = ? AND deleted_at IS NULL",
		listID,
	)
	if err != nil {
		http.Error(w, "Error retrieving list products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item ListProduct
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Checked); err != nil {
			http.Error(w, "Error scanning list product", http.StatusInternalServerError)
			return
		}
		products = append(products, item)
	}

	// Create the response object
	productList := ProductList{
		ID:          int(listID),
		WorkspaceID: workspaceID,
		UserID:      loggedInUserID,
		Title:       req.Title,
		Products:    products,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}

	// Return the created product list
	response := struct {
		Status string      `json:"status"`
		Data   ProductList `json:"data"`
	}{
		Status: "Success",
		Data:   productList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
