package main

import (
	"context"
	"fmt"
	"github.com/gofor-little/env"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type Article struct {
	_id           primitive.ObjectID `bson:"_id"`
	Title         string             `bson:"title"`
	Date          string             `bson:"date"`
	Location      string             `bson:"location"`
	TextPrimary   string             `bson:"textPrimary"`
	TextSecondary string             `bson:"textSecondary"`
}

func main() {
	client := connectToMongoAndReturnInstance()
	router := mux.NewRouter()
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

	router.HandleFunc("/", handlerSample)
	router.HandleFunc("/articles/{id}", getArticle).Methods("GET")

	port, err := getEnvVar("PORT")

	fmt.Println("Listening on port " + port + "...")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func getArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Fprintln(w, "Article request with id:", id)

	client := connectToMongoAndReturnInstance()
	articleCollection := client.Database("test").Collection("articles")
	objectId, objectIdErr := primitive.ObjectIDFromHex(id)
	if objectIdErr != nil {
		log.Fatal(objectIdErr)
	}

	filter := bson.D{{"_id", objectId}}

	var result bson.M
	dataFetchErr := articleCollection.FindOne(context.Background(), filter).Decode(&result)
	if dataFetchErr != nil {
		log.Fatal(dataFetchErr)
	}

	var article Article
	bsonBytes, _ := bson.Marshal(result)
	bson.Unmarshal(bsonBytes, &article)
	tmpl := template.Must(template.ParseFiles("article.html"))
	templateErr := tmpl.Execute(w, article)
	if templateErr != nil {
		log.Fatal(templateErr)
	}

	// write result to response
	//fmt.Fprintln(w, result)
}

func handlerSample(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	films := map[string][]Film{
		"films": {
			{Title: "The Godfather", Director: "Francis Ford Coppola"},
			{Title: "Blade Runner", Director: "Ridley Scott"},
			{Title: "The Thing", Director: "John Carpenter"},
		}}
	tmpl.Execute(w, films)
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
