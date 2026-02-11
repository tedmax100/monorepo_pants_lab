// Order API — a simple HTTP service for managing orders.
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

var orders = []models.Order{
	{ID: "o-1", UserID: "u-1", Product: "Widget", Quantity: 2, Total: 19.98, CreatedAt: time.Now()},
	{ID: "o-2", UserID: "u-2", Product: "Gadget", Quantity: 1, Total: 49.99, CreatedAt: time.Now()},
}

func handleListOrders(w http.ResponseWriter, r *http.Request) {
	httputil.JSONResponse(w, http.StatusOK, orders)
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	order := models.Order{
		ID:        uuid.NewString(),
		UserID:    "u-1",
		Product:   "New Item",
		Quantity:  1,
		Total:     9.99,
		CreatedAt: time.Now(),
	}
	orders = append(orders, order)
	httputil.JSONResponse(w, http.StatusCreated, order)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders", handleListOrders)
	mux.HandleFunc("POST /orders", handleCreateOrder)

	addr := ":8082"
	fmt.Printf("Order API listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
