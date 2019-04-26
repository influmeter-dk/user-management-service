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

// Testing Database Interface methods
func TestDbInterfaceMethods(t *testing.T) {
	testUser := User{
		Account: Account{
			Type:     "email",
			Email:    "test@test.com",
			Password: "testhashedpassword-youcantreadme",
		},
	}

	t.Run("Testing create user", func(t *testing.T) {
		id, err := addUserToDB(testInstanceID, testUser)
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
		_, err := addUserToDB(testInstanceID, testUser)
		if err == nil {
			t.Errorf("user already existed, but created again")
		}
	})

	t.Run("Testing find existing user by id", func(t *testing.T) {
		user, err := getUserByIDFromDB(testInstanceID, testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if user.Account.Email != testUser.Account.Email {
			t.Errorf("found user is not matching test user")
			return
		}
	})

	t.Run("Testing find not existing user by id", func(t *testing.T) {
		_, err := getUserByIDFromDB(testInstanceID, testUser.ID.Hex()+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing find existing user by email", func(t *testing.T) {
		user, err := getUserByEmailFromDB(testInstanceID, testUser.Account.Email)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if user.Account.Email != testUser.Account.Email {
			t.Errorf("found user is not matching test user")
			return
		}
	})

	t.Run("Testing find not existing user by email", func(t *testing.T) {
		_, err := getUserByEmailFromDB(testInstanceID, testUser.Account.Email+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing updating existing user's attributes", func(t *testing.T) {
		testUser.Account.EmailConfirmed = true
		_, err := updateUserInDB(testInstanceID, testUser)
		if err != nil {
			t.Errorf(err.Error())
			return
		}

	})

	t.Run("Testing updating not existing user's attributes", func(t *testing.T) {
		testUser.Account.EmailConfirmed = false
		currentUser := testUser
		id, err := primitive.ObjectIDFromHex(testUser.ID.Hex() + "1")
		currentUser.ID = id
		_, err = updateUserInDB(testInstanceID, currentUser)
		if err == nil {
			t.Errorf("cannot update not existing user")
			return
		}
	})

	t.Run("Testing deleting existing user", func(t *testing.T) {
		err := deleteUserFromDB(testInstanceID, testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("Testing deleting not existing user", func(t *testing.T) {
		err := deleteUserFromDB(testInstanceID, testUser.ID.Hex()+"1")
		if err == nil {
			t.Errorf("user should not be found - error expected")
			return
		}
	})
}
