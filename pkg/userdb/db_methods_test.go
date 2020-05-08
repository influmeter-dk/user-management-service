package userdb

import (
	"testing"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Testing Database Interface methods
func TestDbInterfaceMethods(t *testing.T) {
	testUser := models.User{
		Account: models.Account{
			Type:      "email",
			AccountID: "test@test.com",
			Password:  "testhashedpassword-youcantreadme",
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
		if user.Account.AccountID != testUser.Account.AccountID {
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
		user, err := getUserByEmailFromDB(testInstanceID, testUser.Account.AccountID)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if user.Account.AccountID != testUser.Account.AccountID {
			t.Errorf("found user is not matching test user")
			return
		}
	})

	t.Run("Testing find not existing user by email", func(t *testing.T) {
		_, err := getUserByEmailFromDB(testInstanceID, testUser.Account.AccountID+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing updating existing user's attributes", func(t *testing.T) {
		testUser.Account.AccountConfirmedAt = time.Now().Unix()
		_, err := updateUserInDB(testInstanceID, testUser)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("Testing updating not existing user's attributes", func(t *testing.T) {
		testUser.Account.AccountConfirmedAt = time.Now().Unix()
		currentUser := testUser
		wrongID := testUser.ID.Hex()[:len(testUser.ID.Hex())-2] + "00"
		id, err := primitive.ObjectIDFromHex(wrongID)
		if err != nil {
			t.Error(err)
			return
		}
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
