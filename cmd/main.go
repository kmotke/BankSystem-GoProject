package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type User struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserName      string             `json:"username,omitempty" bson:"username,omitempty"`
	AccountNumber int                `json:"accno,omitempty" bson:"accno,omitempty"`
}

type Account struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AccountNumber int                `json:"accno,omitempty" bson:"accno,omitempty"`
	Balance       int                `json:"balance,omitempty" bson:"balance,omitempty"`
}

var client *mongo.Client
var userColl *mongo.Collection
var accColl *mongo.Collection

func createUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("content-type", "application/json")
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	var checkUser *User
	_ = userColl.FindOne(context.TODO(), bson.M{"accno": user.AccountNumber}).Decode(&checkUser)
	if checkUser != nil {
		str := []interface{}{"User Already exists"}
		json.NewEncoder(w).Encode(str)
		return
	}

	result, _ := userColl.InsertOne(context.TODO(), user)
	accno := Account{
		AccountNumber: user.AccountNumber,
		Balance:       0,
	}

	accno.UserID = result.InsertedID.(primitive.ObjectID)
	_, err := accColl.InsertOne(context.TODO(), accno)

	if err != nil {

		json.NewEncoder(w).Encode(errors.Wrapf(err, "Failed to insert account info"))
		return
	}

	json.NewEncoder(w).Encode(result)

}

func main() {
	//   Connect to my cluster
	var err error

	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin1:root@project1.5wwox.mongodb.net/banksystem?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	userColl = client.Database("banksystem").Collection("user")
	accColl = client.Database("banksystem").Collection("account")

	router := mux.NewRouter()
	router.HandleFunc("/api/create-user", createUser).Methods("POST")

	fmt.Println("Application running...")
	log.Fatal(http.ListenAndServe(":9090", router))

}
