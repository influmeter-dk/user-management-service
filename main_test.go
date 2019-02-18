package main

import (
	"context"
	"os"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	influenzanet "github.com/Influenzanet/api/dist/go"
	user_api "github.com/Influenzanet/api/dist/go/user-management"
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
		resp, err := s.SignupWithEmail(context.Background(), nil)

		if err.Error() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with empty payload", func(t *testing.T) {
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

// Test login
func TestLogin(t *testing.T) {
	// Create Test User
	testUser := User{
		Email:    "test@test.com",
		Password: hashPassword("SuperSecurePassword"),
		Roles:    []string{"PARTICIPANT"},
	}

	id, err := CreateUser(testUser)
	if err != nil {
		t.Errorf("error creating user for testing login")
		return
	}
	testUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
		return
	}

	s := userManagementServer{}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.LoginWithEmail(context.Background(), nil)
		if err == nil || err.Error() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with empty payload", func(t *testing.T) {
		req := &influenzanet.UserCredentials{}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || err.Error() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with wrong email", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:    "test1@test.com",
			Password: testUser.Password,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || err.Error() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with wrong password", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:    testUser.Email,
			Password: "SuperSecurePassword1",
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || err.Error() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("Testing with valid fields", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:    testUser.Email,
			Password: "SuperSecurePassword",
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

	// db cleanup
	DeleteUser(testUser.ID.Hex())
}

func TestPasswordChange(t *testing.T) {
	oldPassword := "SuperSecurePassword"
	newPassword := "NewSuperSecurePassword"

	// Create Test User
	testUser := User{
		Email:    "test@test.com",
		Password: hashPassword(oldPassword),
		Roles:    []string{"PARTICIPANT"},
	}

	id, err := CreateUser(testUser)
	if err != nil {
		t.Errorf("error creating users for testing pw change")
		return
	}
	testUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
		return
	}

	s := userManagementServer{}

	t.Run("Testing without payload", func(t *testing.T) {
		resp, err := s.ChangePassword(context.Background(), nil)

		if err.Error() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing without auth fields", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{}
		resp, err := s.ChangePassword(context.Background(), req)
		if err.Error() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong user id", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId: "test-wrong-id",
				Roles:  []string{"PARTICIPANT"},
			},
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		if err.Error() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with wrong old password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId: id,
				Roles:  []string{"PARTICIPANT"},
			},
			OldPassword: oldPassword + "wrong",
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		if err.Error() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with too weak new password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId: id,
				Roles:  []string{"PARTICIPANT"},
			},
			OldPassword: oldPassword,
			NewPassword: "short",
		}
		resp, err := s.ChangePassword(context.Background(), req)
		if err.Error() != "new password too weak" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("Testing with valid data and new password", func(t *testing.T) {
		req := &user_api.PasswordChangeMsg{
			Auth: &influenzanet.ParsedToken{
				UserId: id,
				Roles:  []string{"PARTICIPANT"},
			},
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		resp, err := s.ChangePassword(context.Background(), req)
		if err != nil || resp == nil {
			t.Errorf("unexpected error: %s", err.Error())
			t.Errorf("or missing response: %s", resp)
		}

		// Check login with new credentials:
		req2 := &influenzanet.UserCredentials{
			Email:    testUser.Email,
			Password: newPassword,
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

	// TEAR DOWN
	DeleteUser(testUser.ID.Hex())
}
