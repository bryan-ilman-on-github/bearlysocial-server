package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"bearlysocial-backend/internal/database"
)

func main() {
	// Initialize MongoDB.
	database.InitMongoDB()
	defer func() {
		if database.MongoClient != nil {
			if err := database.MongoClient.Disconnect(context.Background()); err != nil {
				fmt.Println("ERROR DISCONNECTING FROM MongoDB:", err)
			}
		}
	}()

	// Start server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "80" // Default port if not specified in environment variables.
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		// Add your handlers or router here.
	}

	fmt.Printf("Starting server on port %s.\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("ERROR STARTING SERVER:", err)
	}
}
