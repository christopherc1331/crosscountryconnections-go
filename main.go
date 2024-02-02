package main

import (
	"context"
	"fmt"
	"github.com/gofor-little/env"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Film struct {
	Title    string
	Director string
}

func main() {
	client := connectToMongoAndReturnInstance()
	// query client and print all items collection called "sanity"
	collection := client.Database("test").Collection("foobar")
	cursor, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	// iterate through the cursor and print each document
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}

	handlerSample := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		films := map[string][]Film{
			"films": {
				{Title: "The Godfather", Director: "Francis Ford Coppola"},
				{Title: "Blade Runner", Director: "Ridley Scott"},
				{Title: "The Thing", Director: "John Carpenter"},
			}}
		tmpl.Execute(w, films)
	}
	http.HandleFunc("/", handlerSample)

	port, err := getEnvVar("PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func getEnvVar(key string) (string, error) {
	connectionURI := os.Getenv(key)
	if connectionURI == "" {

		// Load an .env file and set the key-value pairs as environment variables.
		if err := env.Load(".env"); err != nil {
			panic(err)
		}

		valFromDotEnv, err := env.MustGet(key)
		if err != nil {
			log.Fatal(fmt.Sprintf("%s is not set", key))
		}

		return valFromDotEnv, err
	}

	return connectionURI, nil
}

func connectToMongoAndReturnInstance() *mongo.Client {
	fmt.Println("Connecting to db...")

	connectionURI, err := getEnvVar("MONGO_URL")

	// Set client options
	clientOptions := options.Client().ApplyURI(connectionURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the MongoDB server to check if the connection is successful
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}
