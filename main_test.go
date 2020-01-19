package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var testInstanceID = strconv.FormatInt(time.Now().Unix(), 10)

func dropTestDB() {
	log.Println("Drop test database")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbClient.Database(conf.DB.DBNamePrefix + testInstanceID + "_users").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func addTestUsers(userDefs []User) (users []User, err error) {
	for _, uc := range userDefs {
		ID, err := addUserToDB(testInstanceID, uc)
		if err != nil {
			return users, err
		}
		_id, _ := primitive.ObjectIDFromHex(ID)
		uc.ID = _id
		users = append(users, uc)
	}
	return
}

// Pre-Test Setup
func TestMain(m *testing.M) {
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}
