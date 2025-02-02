package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"bearlysocial-backend/api/handler"
	"bearlysocial-backend/util"
)

func main() {
	// Initialize environment.
	util.LoadEnv()

	// Initialize MongoDB.
	util.InitMongoDB()
	defer func() {
		if util.MongoClient != nil {
			if err := util.MongoClient.Disconnect(context.Background()); err != nil {
				fmt.Println("ERROR DISCONNECTING FROM MongoDB:", err)
			}
		}
	}()

	// Set up routing.
	http.HandleFunc("/request-otp", handler.RequestOTP)

	// Start server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "80" // Default port if not specified in environment variables.
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	fmt.Printf("Starting server on port %s.\n", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("ERROR STARTING SERVER:", err)
	}
}
