package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"shopping_list/auth"
	"shopping_list/db"
	"shopping_list/lists"
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

	// Users routes
	r.HandleFunc("/users/register", auth.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", auth.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/logout", auth.Logout).Methods(http.MethodPost)

	// Workspaces routes
	r.HandleFunc("/workspaces", middleware.TokenAuthMiddleware(workspaces.CreateWorkspace)).Methods(http.MethodPost)
	r.HandleFunc("/workspaces", middleware.TokenAuthMiddleware(workspaces.ListWorkspaces)).Methods(http.MethodGet)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.UpdateWorkspace)).Methods(http.MethodPatch)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.DeleteWorkspace)).Methods(http.MethodDelete)
	r.HandleFunc("/workspaces/{workspace_id}", middleware.TokenAuthMiddleware(workspaces.GetWorkspace)).Methods(http.MethodGet)
	r.HandleFunc("/workspaces/{workspace_id}/add_user/{user_id}", middleware.TokenAuthMiddleware(workspaces.AddUserToWorkspace)).Methods(http.MethodPost)
	r.HandleFunc("/workspaces/{workspace_id}/remove_user/{user_id}", middleware.TokenAuthMiddleware(workspaces.RemoveUserFromWorkspace)).Methods(http.MethodDelete)
	r.HandleFunc("/workspaces/{workspace_id}/users", middleware.TokenAuthMiddleware(workspaces.ListUsersInWorkspace)).Methods(http.MethodGet)

	// Products routes
	r.HandleFunc("/workspaces/{workspace_id}/products", middleware.TokenAuthMiddleware(middleware.CombinedWorkspaceMiddleware(products.ProductsHandler)))
	r.HandleFunc("/workspaces/{workspace_id}/products/{id}", middleware.TokenAuthMiddleware(middleware.CombinedWorkspaceMiddleware(products.ProductHandler)))

	// Product Lists routes
	r.HandleFunc("/workspaces/{workspace_id}/product-lists", middleware.TokenAuthMiddleware(middleware.WorkspaceMiddleware(lists.CreateProductList))).Methods(http.MethodPost)
	r.HandleFunc("/workspaces/{workspace_id}/product-lists/{list_id}/status", middleware.TokenAuthMiddleware(middleware.WorkspaceMiddleware(lists.UpdateListStatus))).Methods(http.MethodPatch)
	r.HandleFunc("/workspaces/{workspace_id}/product-lists/{list_id}/products/{product_id}", middleware.TokenAuthMiddleware(middleware.WorkspaceMiddleware(lists.DeleteProductFromList))).Methods(http.MethodDelete)

	err := http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
