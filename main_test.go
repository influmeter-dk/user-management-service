package main

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

// Testing Database Interface methods
func TestDbInterfaceMethods(t *testing.T) {
	testUser := User{
		Email:    "test@test.com",
		Password: "testhashedpassword-youcantreadme",
	}

	t.Run("Testing create user", func(t *testing.T) {
		id, err := CreateUser(testUser)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if len(id) == 0 {
			t.Errorf("id is missing")
			return
		}
		_id, _ := objectid.FromHex(id)
		testUser.ID = _id
	})

	t.Run("Testing creating existing user", func(t *testing.T) {
		_, err := CreateUser(testUser)
		if err == nil {
			t.Errorf("user already existed, but created again")
		}
	})

	t.Run("Testing find existing user by id", func(t *testing.T) {
		user, err := FindUserByID(testUser.ID.Hex())
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
		_, err := FindUserByID(testUser.ID.Hex() + "1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing find existing user by email", func(t *testing.T) {
		user, err := FindUserByEmail(testUser.Email)
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
		_, err := FindUserByEmail(testUser.Email + "1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing updating existing user's attributes", func(t *testing.T) {
		t.Errorf("TEST MISSING")
	})

	t.Run("Testing updating not existing user's attributes", func(t *testing.T) {
		t.Errorf("TEST MISSING")
	})

	t.Run("Testing deleting existing user", func(t *testing.T) {
		t.Errorf("TEST MISSING")
	})

	t.Run("Testing deleting not existing user", func(t *testing.T) {
		t.Errorf("TEST MISSING")
	})
}
