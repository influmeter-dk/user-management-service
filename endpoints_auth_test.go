package main

import (
	"context"
	"testing"

	influenzanet "github.com/influenzanet/api/dist/go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/status"
)

func TestLogin(t *testing.T) {
	s := userManagementServer{}

	// Create Test User
	testUser := User{
		Email:    "test-login@test.com",
		Password: hashPassword("SuperSecurePassword123!§$"),
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

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.LoginWithEmail(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &influenzanet.UserCredentials{}

		resp, err := s.LoginWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("with wrong email", func(t *testing.T) {
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

	t.Run("with wrong password", func(t *testing.T) {
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

	t.Run("with valid fields", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:      testUser.Email,
			Password:   "SuperSecurePassword123!§$",
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

func TestSignup(t *testing.T) {
	s := userManagementServer{}

	wrongEmailFormatNewUserReq := &influenzanet.UserCredentials{
		Email:      "test-signup",
		Password:   "SuperSecurePassword123!§$",
		InstanceId: testInstanceID,
	}

	wrongPasswordFormatNewUserReq := &influenzanet.UserCredentials{
		Email:      "test-signup@test.com",
		Password:   "short",
		InstanceId: testInstanceID,
	}
	validNewUserReq := &influenzanet.UserCredentials{
		Email:      "test-signup@test.com",
		Password:   "SuperSecurePassword123!§$",
		InstanceId: testInstanceID,
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &influenzanet.UserCredentials{}
		resp, err := s.SignupWithEmail(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "email not valid" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong email format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongEmailFormatNewUserReq)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "email not valid" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong password format", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), wrongPasswordFormatNewUserReq)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "password too weak" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with valid fields", func(t *testing.T) {
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

	t.Run("with duplicate user (same email)", func(t *testing.T) {
		req := &influenzanet.UserCredentials{
			Email:      "test-signup-1@test.com",
			Password:   "SuperSecurePassword123!§$",
			InstanceId: testInstanceID,
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

func TestTokenRefreshedEndpoint(t *testing.T) {
	t.Error("test not implemented")
}
