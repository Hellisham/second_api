package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"time"
)

type Products struct {
	ID          int64
	Title       string
	Description string
	Price       int
	Tax         int
	CreatedAt   time.Time
}

var db *pgxpool.Pool

func connDB() {
	var err error
	conn := "postgres://admin:admin@localhost:5432/postgres?sslmode=disable"
	db, err = pgxpool.Connect(context.Background(), conn)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Connected to database")
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query(context.Background(), "select * from products")
	if err != nil {
		http.Error(w, "Error getting products", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var products []Products
	for rows.Next() {
		var product Products
		if err := rows.Scan(&product.ID, &product.Title, &product.Description, &product.Price, &product.Tax, &product.CreatedAt); err != nil {
			http.Error(w, "Error scanning products", http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var product Products
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, " Invalid Product", http.StatusInternalServerError)
	}

	product.CreatedAt = time.Now()
	q := "INSERT INTO products(title, description, price, tax, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = db.QueryRow(context.Background(), q, product.Title, product.Description, product.Price, product.Tax, product.CreatedAt).Scan(&product.ID)
	if err != nil {
		http.Error(w, "Error creating product", http.StatusInternalServerError)
	}

}

func main() {
	connDB()
	defer db.Close()

	http.HandleFunc("/products", getProductsHandler)
	http.HandleFunc("/products/create", createProductHandler)

	port := "8000"
	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
