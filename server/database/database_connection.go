package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DBInstance() *mongo.Client {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	MongoDb := os.Getenv("MONGO_URI")
	if MongoDb == "" {
		log.Fatal("MONGODB_URI not set in environment")
	}
	// Set client options and connect to MongoDB
	clientOptions := options.Client().ApplyURI(MongoDb)
	client, err := mongo.Connect(nil, clientOptions)
	fmt.Println("MongoDB connected successfully", client)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
		return nil
	}
	return client
}

// Create a global MongoDB client instance
var Client *mongo.Client = DBInstance()

func OpenCollection(collectionName string) *mongo.Collection {
	collection := Client.Database("magic-stream-movies").Collection(collectionName)
	if collection == nil {
		return nil
	}
	return collection
}
