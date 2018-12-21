package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// Pre-Test Setup
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

func performRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

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
		_id, _ := primitive.ObjectIDFromHex(id)
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
		testUser.EmailConfirmed = true
		err := UpdateUser(testUser)
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
		err = UpdateUser(currentUser)
		if err == nil {
			t.Errorf("cannot update not existing user")
			return
		}
	})

	t.Run("Testing deleting existing user", func(t *testing.T) {
		err := DeleteUser(testUser.ID.Hex())
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("Testing deleting not existing user", func(t *testing.T) {
		err := DeleteUser(testUser.ID.Hex() + "1")
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})
}

// Test middleware
// TODO: move some tests that only matter to middleware to seperate testfunction

// Test signup
func TestSignup(t *testing.T) {
	r := gin.Default()
	r.POST("/v1/signup", bindUserFromBodyMiddleware(), signupHandl)

	validUser := &User{
		Email:    "test@test.com",
		Password: "SuperSecurePassword",
	}

	t.Run("Testing without payload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/signup", nil)
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		value, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists || value != "payload missing" {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with missing fields", func(t *testing.T) {
		emptyUser := &User{
			Email:    "",
			Password: "",
		}

		payload, _ := json.Marshal(emptyUser)

		req, _ := http.NewRequest("POST", "/v1/signup", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with invalid email field", func(t *testing.T) {
		faultyUser := &User{
			Email:    "asdffsadiidijlkfj.sdf",
			Password: "SuperSecurePassword",
		}

		payload, _ := json.Marshal(faultyUser)

		req, _ := http.NewRequest("POST", "/v1/signup", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with valid fields", func(t *testing.T) {
		payload, _ := json.Marshal(validUser)

		req, _ := http.NewRequest("POST", "/v1/signup", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusCreated || exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
		var err error
		validUser.ID, err = primitive.ObjectIDFromHex(response["_id"])
		if err != nil {
			t.Errorf("error adding ID to valid user\n%s", err)
		}
	})

	t.Run("Testing with existing user", func(t *testing.T) {
		validUser2 := &User{
			Email:    validUser.Email,
			Password: validUser.Password,
		}

		payload, _ := json.Marshal(validUser2)

		req, _ := http.NewRequest("POST", "/v1/signup", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	// db cleanup
	DeleteUser(validUser.ID.Hex())
}

// Test login
func TestLogin(t *testing.T) {
	r := gin.Default()
	r.POST("/v1/signup", bindUserFromBodyMiddleware(), signupHandl)
	r.POST("/v1/login", bindUserFromBodyMiddleware(), loginHandl)

	validUser := User{
		Email:    "test@test.com",
		Password: "SuperSecurePassword",
	}
	hashedUser := User{
		Email:    validUser.Email,
		Password: hashPassword(validUser.Password),
		Roles:    []string{"PARTICIPANT"},
	}

	// db buildup
	id, err := CreateUser(hashedUser)
	if err != nil {
		t.Errorf("error creating user for testing login")
	}
	validUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
	}

	t.Run("Testing without payload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/login", nil)
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		value, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists || value != "payload missing" {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with missing fields", func(t *testing.T) {
		emptyUser := &User{
			Email:    "",
			Password: "",
		}

		payload, _ := json.Marshal(emptyUser)

		req, _ := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusBadRequest || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with wrong email", func(t *testing.T) {
		invalidUser1 := &User{
			Email:    "test@test11.com",
			Password: "SuperSecurePassword",
		}

		payload, _ := json.Marshal(invalidUser1)

		req, _ := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusUnauthorized || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with wrong password", func(t *testing.T) {
		invalidUser2 := &User{
			Email:    "test@test.com",
			Password: "SuperWrongPassword",
		}

		payload, _ := json.Marshal(invalidUser2)

		req, _ := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusUnauthorized || !exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing with valid fields", func(t *testing.T) {
		validUser2 := User{
			Email:    validUser.Email,
			Password: validUser.Password,
		}
		payload, _ := json.Marshal(validUser2)

		req, _ := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]string
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, exists := response["error"]
		if w.Code != http.StatusOK || exists {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	// db cleanup
	DeleteUser(validUser.ID.Hex())
}
