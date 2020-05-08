package userdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var testDBService *UserDBService

const (
	testDBNamePrefix = "TEST_"
)

var (
	testInstanceID = strconv.FormatInt(time.Now().Unix(), 10)
)

func setupTestDBService() {
	connStr := os.Getenv("USER_DB_CONNECTION_STR")
	username := os.Getenv("USER_DB_USERNAME")
	password := os.Getenv("USER_DB_PASSWORD")
	prefix := os.Getenv("USER_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}
	testDBService = NewUserDBService(
		URI, Timeout, IdleConnTimeout, MaxPoolSize, testDBNamePrefix,
	)
}

func dropTestDB() {
	log.Println("Drop test database: userdb package")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := testDBService.DBClient.Database(testDBNamePrefix + testInstanceID + "_users").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// Pre-Test Setup
func TestMain(m *testing.M) {
	setupTestDBService()
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}

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
		id, err := testDBService.AddUser(testInstanceID, testUser)
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
		_, err := testDBService.AddUser(testInstanceID, testUser)
		if err == nil {
			t.Errorf("user already existed, but created again")
		}
	})

	t.Run("Testing find existing user by id", func(t *testing.T) {
		user, err := testDBService.GetUserByID(testInstanceID, testUser.ID.Hex())
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
		_, err := testDBService.GetUserByID(testInstanceID, testUser.ID.Hex()+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing find existing user by email", func(t *testing.T) {
		user, err := testDBService.GetUserByEmail(testInstanceID, testUser.Account.AccountID)
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
		_, err := testDBService.GetUserByEmail(testInstanceID, testUser.Account.AccountID+"1")
		if err == nil {
			t.Errorf("user should not be found")
			return
		}
	})

	t.Run("Testing updating existing user's attributes", func(t *testing.T) {
		testUser.Account.AccountConfirmedAt = time.Now().Unix()
		_, err := testDBService.UpdateUser(testInstanceID, testUser)
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
		_, err = testDBService.UpdateUser(testInstanceID, currentUser)
		if err == nil {
			t.Errorf("cannot update not existing user")
			return
		}
	})

	t.Run("Testing deleting existing user", func(t *testing.T) {
		err := testDBService.DeleteUser(testInstanceID, testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("Testing deleting not existing user", func(t *testing.T) {
		err := testDBService.DeleteUser(testInstanceID, testUser.ID.Hex()+"1")
		if err == nil {
			t.Errorf("user should not be found - error expected")
			return
		}
	})
}
