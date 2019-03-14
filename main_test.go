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

var testInstanceID = "test-db-" + strconv.FormatInt(time.Now().Unix(), 10)

func dropTestDB() {
	log.Println("Drop test database")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbClient.Database(testInstanceID).Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// Pre-Test Setup
func TestMain(m *testing.M) {
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}

// Testing Database Interface methods
func TestDbInterfaceMethods(t *testing.T) {
	testUser := User{
		Email:    "test@test.com",
		Password: "testhashedpassword-youcantreadme",
	}

	t.Run("Testing create user", func(t *testing.T) {
		id, err := createUserDB(testInstanceID, testUser)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if len(id) == 0 {
			t.Errorf("id is missing")
			return
		}
		_id, _ := primitive.ObjectIDFromHex(id)
		testUser.ID = _id
	})

	t.Run("Testing creating existing user", func(t *testing.T) {
		_, err := createUserDB(testInstanceID, testUser)
		if err == nil {
			t.Errorf("user already existed, but created again")
		}
	})

	t.Run("Testing find existing user by id", func(t *testing.T) {
		user, err := findUserByID(testInstanceID, testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if user.Email != testUser.Email {
			t.Errorf("found user is not matching test user")
			return
		}
	})

	t.Run("Testing find not existing user by id", func(t *testing.T) {
		_, err := findUserByID(testInstanceID, testUser.ID.Hex()+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing find existing user by email", func(t *testing.T) {
		user, err := findUserByEmail(testInstanceID, testUser.Email)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if user.Email != testUser.Email {
			t.Errorf("found user is not matching test user")
			return
		}
	})

	t.Run("Testing find not existing user by email", func(t *testing.T) {
		_, err := findUserByEmail(testInstanceID, testUser.Email+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing updating existing user's attributes", func(t *testing.T) {
		testUser.EmailConfirmed = true
		err := updateUserDB(testInstanceID, testUser)
		if err != nil {
			t.Errorf(err.Error())
			return
		}

	})

	t.Run("Testing updating not existing user's attributes", func(t *testing.T) {
		testUser.EmailConfirmed = false
		currentUser := testUser
		id, err := primitive.ObjectIDFromHex(testUser.ID.Hex() + "1")
		currentUser.ID = id
		err = updateUserDB(testInstanceID, currentUser)
		if err == nil {
			t.Errorf("cannot update not existing user")
			return
		}
	})

	t.Run("Testing deleting existing user", func(t *testing.T) {
		err := deleteUserDB(testInstanceID, testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("Testing deleting not existing user", func(t *testing.T) {
		err := deleteUserDB(testInstanceID, testUser.ID.Hex()+"1")
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})
}
