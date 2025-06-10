package main

import (
	"crypto-wallet/pkg/controller/middleware"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	// http routes
	http.HandleFunc("/users", middleware.Users)
	http.HandleFunc("/transactions", middleware.Transactions)

	// start up server
	fmt.Println("server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
