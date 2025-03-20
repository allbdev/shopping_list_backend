package lists

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"shopping_list/db"
	"strconv"

	"github.com/gorilla/mux"
)

type Product struct {
	ListProduct
	Name string `json:"name"`
}

type ProductListWithProducts struct {
	ProductList
	Products []Product
}

// ListProductLists handles listing all product lists in a workspace, including their products and product names
func ListProductLists(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from URL parameters
	vars := mux.Vars(r)
	workspaceID, err := strconv.Atoi(vars["workspace_id"])
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	// Retrieve all product lists and their products in the workspace
	rows, err := db.DB.Query(`
		SELECT l.id, l.user_id, l.title, l.created_at, l.updated_at, l.deleted_at,
		       lp.product_id, lp.quantity, lp.checked, lp.created_at, lp.updated_at, lp.deleted_at,
		       p.title
		FROM lists l
		LEFT JOIN list_products lp ON l.id = lp.list_id AND lp.deleted_at IS NULL
		LEFT JOIN products p ON lp.product_id = p.id AND p.deleted_at IS NULL
		WHERE l.workspace_id = ? AND l.deleted_at IS NULL
		ORDER BY l.id DESC
	`, workspaceID)
	if err != nil {
		http.Error(w, "Error fetching product lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Map to hold lists and their products
	listMap := make(map[int]*ProductListWithProducts)

	for rows.Next() {
		var listID int
		var list ProductListWithProducts
		var product Product
		var productName string

		// Scan list and product details
		// Use nullable types for product fields to handle NULL values
		var productID, quantity sql.NullInt64
		var checked sql.NullBool
		var productCreatedAt, productUpdatedAt, productDeletedAt sql.NullTime
		var nullableProductName sql.NullString

		err := rows.Scan(
			&listID, &list.UserID, &list.Title, &list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			&productID, &quantity, &checked, &productCreatedAt, &productUpdatedAt, &productDeletedAt,
			&nullableProductName,
		)

		// Convert nullable types to regular types if valid
		if productID.Valid {
			product.ProductID = int(productID.Int64)
			product.Quantity = int(quantity.Int64)
			product.Checked = checked.Bool
			product.CreatedAt = productCreatedAt.Time
			product.UpdatedAt = productUpdatedAt.Time
			if productDeletedAt.Valid {
				product.DeletedAt = &productDeletedAt.Time
			}
			if nullableProductName.Valid {
				productName = nullableProductName.String
			}
		}
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error scanning product list", http.StatusInternalServerError)
			return
		}

		// Check if the list is already in the map
		if existingList, exists := listMap[listID]; exists {
			// Add product to existing list
			if product.ProductID != 0 { // Ensure product exists
				product.Name = productName
				product.ListID = listID
				existingList.Products = append(existingList.Products, product)
			}
		} else {

			// Create a new list entry
			list.ID = listID
			if product.ProductID != 0 { // Ensure product exists
				product.Name = productName
				product.ListID = listID
				list.Products = append(list.Products, product)
			}
			listMap[listID] = &list
		}
	}

	// Convert map to slice
	var productLists []ProductListWithProducts
	for _, list := range listMap {
		productLists = append(productLists, *list)
	}

	// Return the list of product lists
	response := struct {
		Status string                    `json:"status"`
		Data   []ProductListWithProducts `json:"data"`
	}{
		Status: "Success",
		Data:   productLists,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
