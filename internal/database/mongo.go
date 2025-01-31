package database

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoCollection *mongo.Collection

// Initialize MongoDB connection.
func InitMongoDB() {
	// Load environment variables from the .env file.
	err := godotenv.Load()
	if err != nil {
		fmt.Println("ERROR LOADING .env FILE:", err)
		os.Exit(1)
	}

	// Get MongoDB credentials from environment variables.
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	collectionName := os.Getenv("MONGO_COLLECTION")

	if mongoURI == "" || dbName == "" || collectionName == "" {
		fmt.Println("MongoDB credentials are missing in .env file.")
		os.Exit(1)
	}

	// Connect to MongoDB.
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("ERROR CONNECTING TO MongoDB:", err)
		os.Exit(1)
	}

	// Ping MongoDB to ensure the connection is established.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println("ERROR PINGING MongoDB:", err)
		os.Exit(1)
	}

	MongoClient = client
	MongoCollection = client.Database(dbName).Collection(collectionName)
	fmt.Println("Connected to MongoDB.")
}
