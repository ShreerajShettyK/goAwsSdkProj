package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
	var err error

	// Load environment variables from .env file
	err = godotenv.Load("/root/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("mongoDbConnectionString")
	if mongoURI == "" {
		log.Fatal("MONGODB connection string not set")
	}
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDb Server started... on 8000")
	http.HandleFunc("/", fetchData)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func fetchData(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("mydatabase").Collection("imagetags")

	filter := bson.D{}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, result := range results {
		fmt.Fprintf(w, "%v\n", result)
	}
}
