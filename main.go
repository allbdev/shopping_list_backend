package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"shopping_list/auth"
	"shopping_list/db"
	"shopping_list/middleware"
	"shopping_list/products"

	"github.com/gorilla/mux"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}

func main() {
	db.DbConnect()

	r := mux.NewRouter()
	r.HandleFunc("/", getRoot)

	// Products	routes
	r.HandleFunc("/products", middleware.TokenAuthMiddleware(products.ProductsHandler))
	r.HandleFunc("/products/{id}", middleware.TokenAuthMiddleware(products.ProductHandler))

	// Auth routes
	r.HandleFunc("/users/register", auth.RegisterHandler)
	r.HandleFunc("/users/login", auth.LoginHandler)

	err := http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
