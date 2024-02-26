package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gofor-little/env"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type Film struct {
	Title    string
	Director string
}

type Article struct {
	ObjectId      primitive.ObjectID `bson:"_id"`
	Type          string             `bson:"type"`
	Id            string             `bson:"id"`
	Img           string             `bson:"img"`
	Author        string             `bson:"author"`
	Rank          int                `bson:"rank"`
	Categories    []string           `bson:"categories"`
	Title         string             `bson:"title"`
	Date          string             `bson:"date"`
	Location      string             `bson:"location"`
	TextPrimary   string             `bson:"textPrimary"`
	TextSecondary string             `bson:"textSecondary"`
}

var client *mongo.Client

func main() {
	client = connectToMongoAndReturnInstance()
	router := mux.NewRouter()

	collection := client.Database("test").Collection("foobar")
	cursor, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}

	router.Handle("/static/{file:.*}", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/static/js/{file:.*}", http.StripPrefix("/static/js/", http.FileServer(http.Dir("static/js"))))
	router.Handle("/static/media/{file:.*}", http.StripPrefix("/static/media/", http.FileServer(http.Dir("static/media"))))
	router.Handle("/static/css/{file:.*}", http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))
	router.Handle("/static/css/font-awesome/css/{file:.*}", http.StripPrefix("/static/css/font-awesome/css/", http.FileServer(http.Dir("static/css/font-awesome/css"))))
	router.Handle("/static/css/font-awesome/fonts/{file:.*}", http.StripPrefix("/static/css/font-awesome/fonts/", http.FileServer(http.Dir("static/css/font-awesome/fonts"))))
	router.Handle("/static/fonts/librebaskerville/{file:.*}", http.StripPrefix("/static/fonts/librebaskerville/", http.FileServer(http.Dir("static/fonts/librebaskerville"))))
	router.Handle("/static/fonts/metropolis/{file:.*}", http.StripPrefix("/static/fonts/metropolis/", http.FileServer(http.Dir("static/fonts/metropolis"))))
	router.Handle("/static/images/{file:.*}", http.StripPrefix("/static/images/", http.FileServer(http.Dir("static/images"))))
	router.Handle("/static/images/avatars/{file:.*}", http.StripPrefix("/static/images/avatars/", http.FileServer(http.Dir("static/images/avatars"))))
	router.Handle("/static/images/icons/{file:.*}", http.StripPrefix("/static/images/icons/", http.FileServer(http.Dir("static/images/icons"))))
	router.Handle("/static/images/icons/png/{file:.*}", http.StripPrefix("/static/images/icons/png/", http.FileServer(http.Dir("static/images/icons/png"))))
	router.Handle("/static/images/mejs/{file:.*}", http.StripPrefix("/static/images/mejs/", http.FileServer(http.Dir("static/images/mejs"))))
	router.Handle("/static/images/thumbs/{file:.*}", http.StripPrefix("/static/images/thumbs/", http.FileServer(http.Dir("static/images/thumbs"))))
	router.Handle("/static/images/thumbs/about/{file:.*}", http.StripPrefix("/static/images/thumbs/about/", http.FileServer(http.Dir("static/images/thumbs/about"))))
	router.Handle("/static/images/thumbs/featured/{file:.*}", http.StripPrefix("/static/images/thumbs/featured/", http.FileServer(http.Dir("static/images/thumbs/featured"))))
	router.Handle("/static/images/thumbs/small/{file:.*}", http.StripPrefix("/static/images/thumbs/small/", http.FileServer(http.Dir("static/images/thumbs/small"))))
	router.Handle("/static/images/thumbs/masonry/{file:.*}", http.StripPrefix("/static/images/thumbs/masonry/", http.FileServer(http.Dir("static/images/thumbs/masonry"))))
	router.Handle("/static/images/thumbs/masonry/gallery/{file:.*}", http.StripPrefix("/static/images/thumbs/masonry/gallery/", http.FileServer(http.Dir("static/images/thumbs/masonry/gallery"))))
	router.Handle("/static/images/thumbs/single/{file:.*}", http.StripPrefix("/static/images/thumbs/single/", http.FileServer(http.Dir("static/images/thumbs/single"))))
	router.Handle("/static/images/thumbs/single/audio/{file:.*}", http.StripPrefix("/static/images/thumbs/single/audio/", http.FileServer(http.Dir("static/images/thumbs/single/audio"))))
	router.Handle("/static/images/thumbs/single/gallery/{file:.*}", http.StripPrefix("/static/images/thumbs/single/gallery/", http.FileServer(http.Dir("static/images/thumbs/single/gallery"))))
	router.Handle("/static/images/thumbs/single/standard/{file:.*}", http.StripPrefix("/static/images/thumbs/single/standard/", http.FileServer(http.Dir("static/images/thumbs/single/standard"))))

	router.HandleFunc("/", getIndex)
	router.HandleFunc("/sample", handlerSample)
	router.HandleFunc("/articles/{id}", getArticleById).Methods("GET")
	router.HandleFunc("/404", get404).Methods("GET")

	// Add a custom 404 handler
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request path
		log.Printf("404 Not Found: %s", r.URL.Path)

		tmpl := template.Must(template.ParseFiles("./static/html/404.html"))
		tmpl.Execute(w, nil)
	})

	printRoutes(router)

	// Listen for an interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		// Disconnect the client when an interrupt signal is received
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	port, err := getEnvVar("PORT")
	fmt.Println("Listening on port " + port + "...")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func get404(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/html/404.html"))
	tmpl.Execute(w, nil)
}

func appendTemplateVars(templateVarsMap map[string]interface{}, articles map[string]interface{}) map[string]interface{} {
	for k, v := range articles {
		templateVarsMap[k] = v
	}
	return templateVarsMap
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	templateVarsMap := make(map[string]interface{})
	highlightedArticles, err := getHighlightedArticlesByRank()
	if err != nil {
		log.Fatal(err)
	}
	// add the highlighted articles to the templateVarsMap
	templateVarsMap = appendTemplateVars(templateVarsMap, highlightedArticles)

	articleCards, err := getArticleCardsOrderedByDate()
	if err != nil {
		log.Fatal(err)
	}
	// add the article cards to the templateVarsMap
	templateVarsMap = appendTemplateVars(templateVarsMap, articleCards)

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	templateErr := tmpl.Execute(w, templateVarsMap)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getArticleById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Println("Article request with id:", id)

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
	article.Id = article.ObjectId.Hex()
	tmpl := template.Must(template.ParseFiles("./templates/article.html"))
	templateErr := tmpl.Execute(w, article)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getHighlightedArticlesByRank() (map[string]interface{}, error) {
	articleCollection := client.Database("test").Collection("articles")

	// Create a cursor for the query
	// filter out articles with a rank of 0
	opts := options.Find().SetSort(bson.D{{"rank", 1}})
	cursor, err := articleCollection.Find(context.Background(), bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	articles := make(map[string]interface{})
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		var article Article
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &article)
		article.Id = article.ObjectId.Hex()
		log.Println(article.Title, article.Rank)

		tmpl := template.Must(template.ParseFiles("./templates/fractional/highlighted.html"))
		var tpl bytes.Buffer
		templateErr := tmpl.Execute(&tpl, article)
		if templateErr != nil {
			return nil, templateErr
		}

		articles[fmt.Sprintf("Highlighted%d", article.Rank)] = template.HTML(tpl.String())
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func getArticleCardsOrderedByDate() (map[string]interface{}, error) {
	articleCollection := client.Database("test").Collection("articles")

	// Create a cursor for the query
	opts := options.Find().SetSort(bson.D{{"date", 1}})
	cursor, err := articleCollection.Find(context.Background(), bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	articlesArray := make([]interface{}, 0)
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}

		var article Article
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &article)
		article.Id = article.ObjectId.Hex()

		// Skip articles without a rank
		if article.Rank != 0 {
			continue
		}

		var tmpl *template.Template
		switch article.Type {
		case "standard":
			tmpl = template.Must(template.ParseFiles("./templates/fractional/article-card-standard.html"))
		case "quote":
			tmpl = template.Must(template.ParseFiles("./templates/fractional/article-card-quote.html"))
		case "gallery":
			tmpl = template.Must(template.ParseFiles("./templates/fractional/article-card-gallery.html"))
		case "audio":
			tmpl = template.Must(template.ParseFiles("./templates/fractional/article-card-audio.html"))
		case "video":
			tmpl = template.Must(template.ParseFiles("./templates/fractional/article-card-video.html"))
		}
		var tpl bytes.Buffer
		templateErr := tmpl.Execute(&tpl, article)
		if templateErr != nil {
			return nil, templateErr
		}

		articlesArray = append(articlesArray, template.HTML(tpl.String()))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	articles := make(map[string]interface{})
	articles["ArticleCards"] = articlesArray

	return articles, nil

}

func handlerSample(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("sample.html"))
	films := map[string][]Film{
		"films": {
			{Title: "The Godfather", Director: "Francis Ford Coppola"},
			{Title: "Blade Runner", Director: "Ridley Scott"},
			{Title: "The Thing", Director: "John Carpenter"},
		}}
	tmpl.Execute(w, films)
}

func getEnvVar(key string) (string, error) {
	envVarVal := os.Getenv(key)
	if envVarVal == "" {

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

	return envVarVal, nil
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

func printRoutes(router *mux.Router) {
	fmt.Println("Available routes:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Printf("Route: %s\n", pathTemplate)
		}
		return nil
	})
}
