package products

import "net/http"

type Product struct {
	ID        int     `json:"id"`
	Title     string  `json:"title"`
	AmoutType string  `json:"amount_type"`
	Price     float32 `json:"price"`
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ListProducts(w, r)
	case http.MethodPost:
		CreateProdcut(w, r)
	case http.MethodPut:
		// Update an existing record.
	case http.MethodDelete:
		// Remove the record.
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
