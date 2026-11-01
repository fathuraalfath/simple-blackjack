package main

import (
	"log"
	"os"

	"simple-blackjack/config"
	"simple-blackjack/routers"
)

func main() {
	// connect db & run migrations
	config.ConnectDatabase()

	// setup router
	r := routers.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
