package main

import (
	"context"
	"os"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	user_api "github.com/Influenzanet/api/user-management"
)

// Pre-Test Setup
func TestMain(m *testing.M) {
	os.Exit(m.Run())
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

// Test signup
func TestSignup(t *testing.T) {
	s := userManagementServer{}

	emptyNewUserReq := &user_api.NewUser{}

	wrongEmailFormatNewUserReq := &user_api.NewUser{
		Email:    "my-email",
		Password: "SuperSecurePassword",
	}

	wrongPasswordFormatNewUserReq := &user_api.NewUser{
		Email:    "test@test.com",
		Password: "short",
	}
	validNewUserReq := &user_api.NewUser{
		Email:    "test@test.com",
		Password: "SuperSecurePassword",
	}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), emptyNewUserReq)

		if err.Error() != "email not valid" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong email format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongEmailFormatNewUserReq)

		if err.Error() != "email not valid" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong password format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongPasswordFormatNewUserReq)

		if err.Error() != "password too weak" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with valid fields", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), validNewUserReq)

		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp == nil || len(resp.UserId) < 3 || len(resp.Roles) < 1 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		DeleteUser(resp.UserId)
	})

	t.Run("Testing signupwith duplicate user (same email)", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), validNewUserReq)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		uID := resp.UserId

		// Try to signup again:
		resp, err = s.SignupWithEmail(context.Background(), validNewUserReq)
		if err == nil || resp != nil {
			t.Errorf("should fail, when user exists already, wrong response: %s", resp)
			return
		}

		DeleteUser(uID)
	})
}

/*
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

		var response map[string]interface{}
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

func TestPasswordChange(t *testing.T) {
	r := gin.Default()
	r.POST("/v1/login", bindUserFromBodyMiddleware(), loginHandl)
	r.POST("/v1/change-password", bindUserFromBodyMiddleware(), passwordChangeHandl)

	// BUILD UP
	testuser := User{
		Email:    "test@test.com",
		Password: hashPassword("testpassword"),
		Roles:    []string{"PARTICIPANT"},
	}

	testuserID, testuserErr := CreateUser(testuser)
	if testuserErr != nil {
		t.Errorf("error creating user for testing login")
	}

	t.Run("Testing without payload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/change-password", nil)
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

		req, _ := http.NewRequest("POST", "/v1/change-password", bytes.NewBuffer(payload))
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
		invalidUser := &User{
			Email:    "testtest@test.com",
			Password: "testpassword",
		}

		payload, _ := json.Marshal(invalidUser)

		req, _ := http.NewRequest("POST", "/v1/change-password", bytes.NewBuffer(payload))
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
		invalidUser := &User{
			Email:    "test@test.com",
			Password: "SuperWrongPassword",
		}

		payload, _ := json.Marshal(invalidUser)

		req, _ := http.NewRequest("POST", "/v1/change-password", bytes.NewBuffer(payload))
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

	t.Run("Testing with non matching new passwords", func(t *testing.T) {
		invalidUser := &User{
			Email:             "test@test.com",
			Password:          "testpassword",
			NewPassword:       "newtestpassword",
			NewPasswordRepeat: "newtestpassword1",
		}

		payload, _ := json.Marshal(invalidUser)

		req, _ := http.NewRequest("POST", "/v1/change-password", bytes.NewBuffer(payload))
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
		validUser := &User{
			Email:             "test@test.com",
			Password:          "testpassword",
			NewPassword:       "newtestpassword",
			NewPasswordRepeat: "newtestpassword",
		}

		payload, _ := json.Marshal(validUser)

		req, _ := http.NewRequest("POST", "/v1/change-password", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(w.Body.String()), &response); err != nil {
			t.Errorf("error parsing response body: %s", err.Error())
		}

		_, errExists := response["error"]
		value, valueExists := response["success"].(bool)
		if w.Code != http.StatusOK || errExists || !valueExists || !value {
			t.Errorf("status code: %d", w.Code)
			t.Errorf("response content: %s", w.Body.String())
			return
		}
	})

	t.Run("Testing login with new password", func(t *testing.T) {
		validUser := &User{
			Email:    "test@test.com",
			Password: "newtestpassword",
		}

		payload, _ := json.Marshal(validUser)

		req, _ := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		w := performRequest(r, req)

		var response map[string]interface{}
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

	// TEAR DOWN
	DeleteUser(testuserID)
}
*/
