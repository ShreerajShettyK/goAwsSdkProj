package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
	var err error

	// Load environment variables from .env file
	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	// mongoURI := os.Getenv("mongoDbConnectionString")
	// if mongoURI == "" {
	// 	log.Fatal("MONGODB connectionstring not set")
	// }
	clientOptions := options.Client().ApplyURI("mongodb+srv://admin:admin@cluster0.0elhpdy.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
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

func insertDummyData() {
	collection := client.Database("TestingDatabase").Collection("Table1")

	// Define dummy data
	dummyData := []interface{}{
		bson.D{{Key: "name", Value: "John Doe"}, {Key: "age", Value: 30}, {Key: "city", Value: "New York"}},
		bson.D{{Key: "name", Value: "Jane Smith"}, {Key: "age", Value: 25}, {Key: "city", Value: "London"}},
		bson.D{{Key: "name", Value: "Bob Johnson"}, {Key: "age", Value: 35}, {Key: "city", Value: "Paris"}},
	}

	// Insert dummy data
	_, err := collection.InsertMany(context.Background(), dummyData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Dummy data inserted successfully")
}
