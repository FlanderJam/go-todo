package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kwilmot/go-todo/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var clientOptions = options.Client().ApplyURI(os.Getenv("MONGO_URI"))

type Todo struct {
	Id primitive.ObjectID `bson:"_id" json:"id"`
	Title string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
	IsComplete bool `bson:"is_complete" json:"is_complete"`
	CreateDate primitive.DateTime `bson:"create_date" json:"create_date"`
	ModifiedDate primitive.DateTime `bson:"modified_date" json:"modified_date"`
	DeleteDate int `bson:"delete_date" json:"delete_date"`

}

type TodoHandler struct {

}

func (h *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var id string
	id, r.URL.Path = utils.ShiftPath(r.URL.Path)
	if id == "" {
		// ALL TODOS
		switch r.Method {
		case "GET":
			h.handleGetAll(w)
		case "POST":
			h.handlePost(w, r)
		default:
			http.Error(w, r.Method + " Method is not allowed!", http.StatusMethodNotAllowed)
		}
	} else {
		// Specific todo
		switch r.Method {
		case "GET":
			h.handleGet(w, id)
		case "PUT":
			h.handlePut(w, r, id)
		case "DELETE":
			h.handleDelete(w, id)
		default:
			http.Error(w, r.Method + " Method is not allowed!", http.StatusMethodNotAllowed)
		}
	}
}

func (h *TodoHandler) handleGetAll(w http.ResponseWriter) {
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}

	collection := dbClient.Database("go-todo-app").Collection("todos")
	// Pass these options to the Find method
	findOptions := options.Find()
	findOptions.SetLimit(10)

	// Here's an array in which you can store the decoded documents
	var results []*Todo

	// set up find filter
	//filter := bson.D{{Key:"delete_date", Value:bson.D{{Key:"$gt", Value:0}}}}
	filter := bson.D{{Key:"delete_date", Value:-1}}
	// Passing bson.D{{}} as the filter matches all documents in the collection
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem Todo
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	err = cursor.Close(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	h.successResponse(w)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		log.Fatal(err)
	}
	// This code can be used to close a connection to the db
	err = dbClient.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
}
func (h *TodoHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var newTodo *Todo
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(requestBody, &newTodo); err != nil {
		log.Fatal(err)
	}
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	postBody := bson.D{
		{Key:"title", Value:newTodo.Title},
		{Key:"description", Value:newTodo.Description},
		{Key:"is_complete", Value:newTodo.IsComplete},
		{Key:"create_date", Value:time.Now()},
		{Key:"modified_date", Value:time.Now()},
		{Key:"delete_date", Value:-1},
	}
	collection := dbClient.Database("go-todo-app").Collection("todos")
	insertResult, err := collection.InsertOne(context.TODO(), postBody)
	if err != nil {
		log.Fatal(err)
	}

	h.successResponse(w)
	err = json.NewEncoder(w).Encode(insertResult)
	if err != nil {
		log.Fatal(err)
	}
	// This code can be used to close a connection to the db
	err = dbClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

}

func (h *TodoHandler) handleGet(w http.ResponseWriter, id string) {
	docId, err := primitive.ObjectIDFromHex(id)
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	collection := dbClient.Database("go-todo-app").Collection("todos")
	var result *Todo
	filter := bson.M{"_id": docId}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	h.successResponse(w)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Fatal(err)
	}
	// This code can be used to close a connection to the db
	err = dbClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
func (h *TodoHandler) handlePut(w http.ResponseWriter, r *http.Request, id string) {
	var updateTodo *Todo
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(requestBody, &updateTodo); err != nil {
		log.Fatal(err)
	}
	putBody := bson.D{
		{Key:"title", Value:updateTodo.Title},
		{Key:"description", Value:updateTodo.Description},
		{Key:"is_complete", Value:updateTodo.IsComplete},
		{Key:"create_date", Value:time.Now()},
		{Key:"modified_date", Value:time.Now()},
		{Key:"delete_date", Value:-1},
	}

	docId, err := primitive.ObjectIDFromHex(id)
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	collection := dbClient.Database("go-todo-app").Collection("todos")
	filter := bson.D{{Key:"_id", Value:docId}}
	update := bson.D{{Key:"$set", Value:putBody}}
	putResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	h.successResponse(w)
	err = json.NewEncoder(w).Encode(putResult)
	if err != nil {
		log.Fatal(err)
	}
	// This code can be used to close a connection to the db
	err = dbClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
func (h *TodoHandler) handleDelete(w http.ResponseWriter, id string) {
	docId, err := primitive.ObjectIDFromHex(id)
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	collection := dbClient.Database("go-todo-app").Collection("todos")
	filter := bson.D{{Key:"_id", Value:docId}}
	update := bson.D{{Key:"$set", Value:bson.D{{Key:"delete_date", Value:time.Now().Unix()}}}}
	deleteResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	h.successResponse(w)
	err = json.NewEncoder(w).Encode(deleteResult)
	if err != nil {
		log.Fatal(err)
	}
	// This code can be used to close a connection to the db
	err = dbClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func (h *TodoHandler) successResponse(w http.ResponseWriter) {
	fmt.Println("setting up successful response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}