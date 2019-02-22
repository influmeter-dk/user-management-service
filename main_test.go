package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"google.golang.org/grpc/status"

	influenzanet "github.com/Influenzanet/api/dist/go"
	user_api "github.com/Influenzanet/api/dist/go/user-management"
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

// Test signup
func TestSignup(t *testing.T) {
	s := userManagementServer{}

	wrongEmailFormatNewUserReq := &user_api.NewUser{
		Auth: &influenzanet.UserCredentials{
			Email:      "test-signup",
			Password:   "SuperSecurePassword",
			InstanceId: testInstanceID,
		},
		Profile: &user_api.Profile{},
	}

	wrongPasswordFormatNewUserReq := &user_api.NewUser{
		Auth: &influenzanet.UserCredentials{
			Email:      "test-signup@test.com",
			Password:   "short",
			InstanceId: testInstanceID,
		},
		Profile: &user_api.Profile{},
	}
	validNewUserReq := &user_api.NewUser{
		Auth: &influenzanet.UserCredentials{
			Email:      "test-signup@test.com",
			Password:   "SuperSecurePassword",
			InstanceId: testInstanceID,
		},
		Profile: &user_api.Profile{},
	}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with empty payload", func(t *testing.T) {
		req := &user_api.NewUser{}
		resp, err := s.SignupWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong email format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongEmailFormatNewUserReq)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "email not valid" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong password format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongPasswordFormatNewUserReq)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "password too weak" || resp != nil {
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
	})

	t.Run("Testing signupwith duplicate user (same email)", func(t *testing.T) {
		req := &user_api.NewUser{
			Auth: &influenzanet.UserCredentials{
				Email:      "test-signup-1@test.com",
				Password:   "SuperSecurePassword",
				InstanceId: testInstanceID,
			},
		}
		resp, err := s.SignupWithEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		// Try to signup again:
		resp, err = s.SignupWithEmail(context.Background(), req)
		if err == nil || resp != nil {
			t.Errorf("should fail, when user exists already, wrong response: %s", resp)
			return
		}
	})
}

// Test login
func TestLogin(t *testing.T) {

	s := userManagementServer{}

	// Create Test User
	testUser := User{
		Email:    "test-login@test.com",
		Password: hashPassword("SuperSecurePassword"),
		Roles:    []string{"PARTICIPANT"},
	}

	id, err := createUserDB(testInstanceID, testUser)
	if err != nil {
		t.Errorf("error creating user for testing login")
		return
	}
	testUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
		return
	}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.LoginWithEmail(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with empty payload", func(t *testing.T) {
		req := &influenzanet.UserCredentials{}

		resp, err := s.LoginWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with wrong email", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:      "test-login@test.com",
			Password:   testUser.Password,
			InstanceId: testInstanceID,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with wrong password", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:      testUser.Email,
			Password:   "SuperSecurePassword1",
			InstanceId: testInstanceID,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with valid fields", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:      testUser.Email,
			Password:   "SuperSecurePassword",
			InstanceId: testInstanceID,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)

		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp == nil || len(resp.UserId) < 3 || len(resp.Roles) < 1 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})
}

func TestPasswordChange(t *testing.T) {
	s := userManagementServer{}

	oldPassword := "SuperSecurePassword"
	newPassword := "NewSuperSecurePassword"

	// Create Test User
	testUser := User{
		Email:    "test-password-change@test.com",
		Password: hashPassword(oldPassword),
		Roles:    []string{"PARTICIPANT"},
	}

	id, err := createUserDB(testInstanceID, testUser)
	if err != nil {
		t.Errorf("error creating users for testing pw change")
		return
	}
	testUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
		return
	}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.ChangePassword(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing without auth fields", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong user id", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId:     "test-wrong-id",
				Roles:      []string{"PARTICIPANT"},
				InstanceId: testInstanceID,
			},
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong old password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId:     id,
				Roles:      []string{"PARTICIPANT"},
				InstanceId: testInstanceID,
			},
			OldPassword: oldPassword + "wrong",
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with too weak new password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId:     id,
				Roles:      []string{"PARTICIPANT"},
				InstanceId: testInstanceID,
			},
			OldPassword: oldPassword,
			NewPassword: "short",
		}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "new password too weak" || resp != nil {
			t.Errorf("wrong error: %s", st.Message())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with valid data and new password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId:     id,
				Roles:      []string{"PARTICIPANT"},
				InstanceId: testInstanceID,
			},
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		if err != nil || resp == nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			t.Errorf("or missing response: %s", resp)
		}

		// Check login with new credentials:
		req2 := &influenzanet.UserCredentials{
			Email:      testUser.Email,
			Password:   newPassword,
			InstanceId: testInstanceID,
		}

		resp2, err := s.LoginWithEmail(context.Background(), req2)

		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp2 == nil || len(resp2.UserId) < 3 || len(resp2.Roles) < 1 {
			t.Errorf("unexpected response: %s", resp2)
			return
		}
	})
}
