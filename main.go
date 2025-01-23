package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB setup
var mongoClient *mongo.Client
var mongoCollection *mongo.Collection

// Initialize MongoDB connection
func initMongoDB() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
	}

	// Get MongoDB credentials from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	collectionName := os.Getenv("MONGO_COLLECTION")

	if mongoURI == "" || dbName == "" || collectionName == "" {
		fmt.Println("MongoDB credentials are missing in .env file.")
		os.Exit(1)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		os.Exit(1)
	}

	// Ping MongoDB to ensure the connection is established
	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println("Error pinging MongoDB:", err)
		os.Exit(1)
	}

	mongoClient = client
	mongoCollection = client.Database(dbName).Collection(collectionName)
	fmt.Println("Connected to MongoDB.")
}

// trackCPU retrieves the actual CPU usage percentage.
func trackCPU() (float64, error) {
	percentages, err := cpu.Percent(0, false) // Get CPU usage for all cores combined.
	if err != nil {
		return 0, err
	}
	return percentages[0], nil // Return the usage of the first core group (all combined).
}

// writeToMongoDB logs CPU usage and timestamp to MongoDB.
func writeToMongoDB(cpuUsage float64) {
	timestamp := time.Now()

	// Create a document to insert
	doc := bson.M{
		"timestamp": timestamp,
		"cpu_usage": cpuUsage,
	}

	// Insert the document into the collection
	_, err := mongoCollection.InsertOne(context.Background(), doc)
	if err != nil {
		fmt.Println("Error inserting document into MongoDB:", err)
		return
	}

	// fmt.Printf("Logged to MongoDB: %v | CPU Usage: %.2f%%\n", timestamp, cpuUsage)
}

// cpuHandler checks the CPU usage and sends it as a response body.
// It also logs the usage to MongoDB.
func cpuHandler(w http.ResponseWriter, r *http.Request) {
	cpuUsage, err := trackCPU()
	if err != nil {
		http.Error(w, "Error retrieving CPU usage", http.StatusInternalServerError)
		return
	}

	// Log CPU usage to MongoDB
	writeToMongoDB(cpuUsage)

	if cpuUsage >= 100.0 {
		http.Error(w, fmt.Sprintf("CPU usage is too high: %.2f%%", cpuUsage), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("Current CPU Usage: %.2f%%", cpuUsage)))
}

func main() {
	// Parse command-line flags
	mode := flag.String("mode", "server", "Mode to run: 'server' or 'bench'.")
	flag.Parse()

	if *mode == "bench" {
		RunBenchmarks()
		os.Exit(0) // Exit after running benchmarks.
	}

	// Initialize MongoDB
	initMongoDB()
	defer func() {
		if mongoClient != nil {
			_ = mongoClient.Disconnect(context.Background())
			fmt.Println("Disconnected from MongoDB.")
		}
	}()

	// Set up the HTTP server
	http.HandleFunc("/cpu", cpuHandler)
	server := &http.Server{
		Addr: ":8080", // Listen on port 8080
	}

	fmt.Println("Starting server on :8080.")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server.")
	}
}
