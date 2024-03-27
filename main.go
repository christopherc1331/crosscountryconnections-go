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
	"sort"
	"strconv"
	"strings"
	"time"
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
	Imgs          []string           `bson:"imgs"`
	Author        string             `bson:"author"`
	Rank          int                `bson:"rank"`
	Categories    []string           `bson:"categories"`
	Title         string             `bson:"title"`
	Date          string             `bson:"date"`
	DateTime      string             `bson:"dateTime"`
	Location      string             `bson:"location"`
	TextPrimary   string             `bson:"textPrimary"`
	TextSecondary string             `bson:"textSecondary"`
	ViewCount     int                `bson:"viewCount"`
	IsArchived    bool               `bson:"isArchived"`
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
	router.HandleFunc("/archives", getArchivesHandler)
	router.HandleFunc("/sample", handlerSample)
	router.HandleFunc("/categories", getCategoriesHandler).Methods("GET")
	router.HandleFunc("/articles/{id}", getArticleById).Methods("GET")
	router.HandleFunc("/articles/highlighted/{type}", getHighlightedArticleHtmlByRankHandler).Methods("GET")
	router.HandleFunc("/articles/popular/{count}", getTopPopularArticlesHandler).Methods("GET")
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

func getIndex(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	category := queryParams.Get("category")
	page, err := strconv.Atoi(queryParams.Get("page"))
	month := queryParams.Get("month")
	year := queryParams.Get("year")
	search := queryParams.Get("s")

	// Store the incremented page value back in the query parameters
	queryParams.Set("page", strconv.Itoa(page))
	r.URL.RawQuery = queryParams.Encode()

	articleCards, err := getArticleCardsOrderedByDate(category, month, year, page, search)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	templateErr := tmpl.Execute(w, articleCards)
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

	// increment the view count
	_, err := articleCollection.UpdateOne(
		context.Background(),
		bson.D{{"_id", objectId}},
		bson.D{{"$inc", bson.D{{"viewCount", 1}}}},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get the query parameters from the request
	queryParams := r.URL.Query()

	// Get the 'type' query parameter
	articleType := queryParams.Get("type")

	// if the 'type' query parameter is set, use the specified template
	var tmpl *template.Template
	if articleType != "" {
		tmpl = template.Must(template.ParseFiles(fmt.Sprintf("./templates/article-%s.html", articleType)))
	} else {
		tmpl = template.Must(template.ParseFiles("./templates/article.html"))
	}
	templateErr := tmpl.Execute(w, article)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getArchivesHandler(w http.ResponseWriter, r *http.Request) {
	articleCollection := client.Database("test").Collection("articles")

	pipeline := mongo.Pipeline{
		{
			{"$addFields", bson.D{
				{"dateObj", bson.D{{"$dateFromString", bson.D{
					{"dateString", "$date"},
					{"format", "%B %d, %Y"},
				}}}},
			}},
		},
		{
			{"$project", bson.D{
				{"monthYear", bson.D{{"$dateToString", bson.D{
					{"format", "%B %Y"},
					{"date", "$dateObj"},
				}}}},
			}},
		},
		{
			{"$group", bson.D{
				{"_id", "$monthYear"},
			}},
		},
		{
			{"$sort", bson.D{
				{"_id", -1},
			}},
		},
		{
			{"$limit", 6},
		},
	}

	cursor, err := articleCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	var archives []string
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		archives = append(archives, result["_id"].(string))
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Split each string in the archives array into Month and Year
	var archivesWithMonthAndYear []map[string]string
	for _, archive := range archives {
		split := strings.Split(archive, " ")
		month := split[0]
		year := split[1]
		archivesWithMonthAndYear = append(archivesWithMonthAndYear, map[string]string{"Month": month, "Year": year})
	}

	archivesMap := make(map[string]interface{})
	archivesMap["Archives"] = archivesWithMonthAndYear

	tmpl := template.Must(template.ParseFiles("./templates/fractional/archives.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templateErr := tmpl.Execute(w, archivesMap)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getTopPopularArticlesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	count, err := strconv.Atoi(vars["count"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	articleCollection := client.Database("test").Collection("articles")

	// Create a cursor for the query
	opts := options.Find().SetSort(bson.D{{"viewCount", -1}}).SetLimit(int64(count))
	filter := bson.D{{"type", "standard"}}
	cursor, err := articleCollection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	// get array of articles, but don't render yet
	articlesArray := make([]Article, 0)
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		var article Article
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &article)
		article.Id = article.ObjectId.Hex()
		// convert date value into datetime value from  Dec 12, 2017 to 2017-12-12
		month := article.Date[0:3]
		day := article.Date[4:6]
		year := article.Date[7:11]

		// Check if day is a single digit and add a leading zero if necessary
		if len(day) == 1 {
			day = "0" + day
		}

		article.DateTime = year + "-" + month + "-" + day

		articlesArray = append(articlesArray, article)
	}

	// create a map of articles
	articles := make(map[string]interface{})
	articles["PopularArticles"] = articlesArray

	// render the articles
	tmpl := template.Must(template.ParseFiles("./templates/fractional/popular-articles.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templateErr := tmpl.Execute(w, articles)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	articleCollection := client.Database("test").Collection("articles")

	categories, err := articleCollection.Distinct(context.Background(), "categories", bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	// put categories in a map
	categoriesMap := make(map[string]interface{})
	categoriesMap["Categories"] = categories

	// get query param for type
	queryParams := r.URL.Query()
	articleType := queryParams.Get("type")

	// if the 'type' query parameter is dropdown, use the specified template
	var tmpl *template.Template
	if articleType == "drop-down" {
		tmpl = template.Must(template.ParseFiles("./templates/fractional/category-drop-down.html"))
	} else {
		tmpl = template.Must(template.ParseFiles("./templates/fractional/category-cloud.html"))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templateErr := tmpl.Execute(w, categoriesMap)
	if templateErr != nil {
		log.Fatal(templateErr)
	}
}

func getHighlightedArticleHtmlByRankHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rankType := vars["type"]

	html, err := getHighlightedArticleHtmlByRank(rankType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func getHighlightedArticleHtmlByRank(rankType string) (template.HTML, error) {
	// Connect to the MongoDB collection
	articleCollection := client.Database("test").Collection("articles")

	var filter bson.D
	if rankType == "primary" {
		filter = bson.D{{"rank", 1}}
	} else {
		filter = bson.D{{"rank", bson.D{{"$gt", 1}, {"$lt", 4}}}}
	} // Query the collection to find all articles
	var results []Article
	cursor, err := articleCollection.Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		var article Article
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &article)
		article.Id = article.ObjectId.Hex()
		results = append(results, article)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles("./templates/fractional/highlighted.html"))

	// Execute the template with the article as the data to be rendered
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, results)
	if err != nil {
		return "", err
	}

	// Return the rendered HTML as a string
	return template.HTML(tpl.String()), nil
}

func getArticleCardsOrderedByDate(category string, month string, year string, page int, search string) (map[string]interface{}, error) {
	pageSize := 13
	articleCollection := client.Database("test").Collection("articles")

	filter := bson.M{}

	if category != "" {
		filter["categories"] = category
	}
	if month != "" && year != "" {
		filter["date"] = bson.M{"$regex": fmt.Sprintf("^%s.*%s", month, year)}
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
			{"textPrimary": bson.M{"$regex": search, "$options": "i"}},
			{"textSecondary": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	filter["isArchived"] = bson.M{"$ne": true}

	// Calculate the number of documents to skip
	skip := int64(page * pageSize)

	if skip < 0 {
		skip = 0
	}

	// Create a cursor for the query
	opts := options.Find().SetSort(bson.D{{"date", 1}}).SetSkip(skip).SetLimit(int64(pageSize))
	cursor, err := articleCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	articlesArray := make([]Article, 0)
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

		articlesArray = append(articlesArray, article)
	}

	// Sort the articles by date
	sort.Slice(articlesArray, func(i, j int) bool {
		dateI, _ := time.Parse("January 2, 2006", articlesArray[i].Date)
		dateJ, _ := time.Parse("January 2, 2006", articlesArray[j].Date)
		return dateI.After(dateJ)
	})

	// Render the template for each sorted article and store the rendered HTML
	renderedArticlesArray := make([]interface{}, 0)
	for _, article := range articlesArray {
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

		renderedArticlesArray = append(renderedArticlesArray, template.HTML(tpl.String()))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	articles := make(map[string]interface{})
	articles["ArticleCards"] = renderedArticlesArray

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
