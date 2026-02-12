// User API — a simple HTTP service for managing users.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/example/monorepo-demo/go/pkg/httputil"
	"github.com/example/monorepo-demo/go/pkg/models"
	"github.com/google/uuid"
)

var users = []models.User{
	{ID: "u-1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()},
	{ID: "u-2", Name: "Bob_test", Email: "bob@example.com", CreatedAt: time.Now()},
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	httputil.JSONResponse(w, http.StatusOK, users)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	user := models.User{
		ID:        uuid.NewString(),
		Name:      "New User",
		Email:     "new@example.com",
		CreatedAt: time.Now(),
	}
	users = append(users, user)
	httputil.JSONResponse(w, http.StatusCreated, user)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", handleListUsers)
	mux.HandleFunc("POST /users", handleCreateUser)

	addr := ":8081"
	fmt.Printf("User API listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
