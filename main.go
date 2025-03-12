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
	"shopping_list/workspaces"

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

	// Products routes
	r.HandleFunc("/products", middleware.TokenAuthMiddleware(products.ProductsHandler))
	r.HandleFunc("/products/{id}", middleware.TokenAuthMiddleware(products.ProductHandler))

	// Users routes
	r.HandleFunc("/users/register", auth.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", auth.Login).Methods(http.MethodPost)

	// Workspaces routes
	r.HandleFunc("/workspaces", middleware.TokenAuthMiddleware(workspaces.CreateWorkspace)).Methods(http.MethodPost)
	r.HandleFunc("/workspaces", middleware.TokenAuthMiddleware(workspaces.ListWorkspaces)).Methods(http.MethodGet)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.UpdateWorkspace)).Methods(http.MethodPatch)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.DeleteWorkspace)).Methods(http.MethodDelete)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.GetWorkspace)).Methods(http.MethodGet)

	err := http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
