package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"bearlysocial-backend/api/handler"
	"bearlysocial-backend/api/middleware"
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

	// Public endpoints for requesting and validating one-time passwords.
	http.HandleFunc("/request-otp", handler.RequestOTP)
	http.HandleFunc("/validate-otp", handler.ValidateOTP)

	// Protected endpoints that require a valid token for access.
	http.Handle("/update-session", middleware.ValidateToken(http.HandlerFunc(handler.UpdateSession)))
	// Others...

	// Benchmark endpoint for performance testing and diagnostics.
	http.HandleFunc("/benchmark", handler.Benchmark)

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
