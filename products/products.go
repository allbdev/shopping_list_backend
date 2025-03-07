package products

import (
	"net/http"
	"time"
)

type Product struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	AmoutType string     `json:"amount_type"`
	Price     float32    `json:"price"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ListProducts(w, r)
	case http.MethodPost:
		CreateProdcut(w, r)
	default:
		http.Error(w, "Method not allowed 1", http.StatusMethodNotAllowed)
	}
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		DeleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed 2", http.StatusMethodNotAllowed)
	}
}
