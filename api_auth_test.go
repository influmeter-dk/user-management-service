package main

import (
	"context"
	"testing"
	"time"

	api "github.com/influenzanet/user-management-service/api"
	"github.com/influenzanet/user-management-service/models"
	utils "github.com/influenzanet/user-management-service/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/status"
)

func TestLogin(t *testing.T) {
	s := userManagementServer{}

	// Create Test User
	currentPw := "SuperSecurePassword123!§$"
	hashedPw, err := utils.HashPassword(currentPw)
	if err != nil {
		t.Errorf("error creating user for testing login")
		return
	}

	testUser := models.User{
		Account: models.Account{
			Type:               "email",
			AccountID:          "test-login@test.com",
			AccountConfirmedAt: time.Now().Unix(),
			Password:           hashedPw,
			PreferredLanguage:  "de",
		},
		Roles: []string{"PARTICIPANT"},
		Profiles: []models.Profile{
			{ID: primitive.NewObjectID()},
		},
	}

	id, err := addUserToDB(testInstanceID, testUser)
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
		req := &api.LoginWithEmailMsg{}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || status.Convert(err).Message() != "invalid username and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
			return
		}
	})

	t.Run("with wrong email", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      "wrong@test.com",
			Password:   currentPw,
			InstanceId: testInstanceID,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "invalid username and/or password" {
			t.Errorf("wrong error: %s", err.Error())
			return
		}
	})

	t.Run("with wrong password", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      testUser.Account.AccountID,
			Password:   currentPw + "w",
			InstanceId: testInstanceID,
		}

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "invalid username and/or password" {
			t.Errorf("wrong error: %s", err.Error())
			return
		}
	})

	t.Run("with valid fields", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      testUser.Account.AccountID,
			Password:   currentPw,
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

		if resp.PreferredLanguage != "de" || !resp.AccountConfirmed {
			t.Errorf("unexpected PreferredLanguage or AccountConfirmed: %s", resp)
			return
		}
	})
}

func TestSignup(t *testing.T) {
	s := userManagementServer{}

	wrongEmailFormatNewUserReq := &api.SignupWithEmailMsg{
		Email:             "test-signup",
		Password:          "SuperSecurePassword123!§$",
		InstanceId:        testInstanceID,
		PreferredLanguage: "en",
	}

	wrongPasswordFormatNewUserReq := &api.SignupWithEmailMsg{
		Email:             "test-signup@test.com",
		Password:          "short",
		InstanceId:        testInstanceID,
		PreferredLanguage: "en",
	}
	validNewUserReq := &api.SignupWithEmailMsg{
		Email:             "test-signup@test.com",
		Password:          "SuperSecurePassword123!§$",
		InstanceId:        testInstanceID,
		PreferredLanguage: "en",
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.SignupWithEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{}
		_, err := s.SignupWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "email not valid")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong email format", func(t *testing.T) {
		_, err := s.SignupWithEmail(context.Background(), wrongEmailFormatNewUserReq)
		ok, msg := shouldHaveGrpcErrorStatus(err, "email not valid")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong password format", func(t *testing.T) {
		_, err := s.SignupWithEmail(context.Background(), wrongPasswordFormatNewUserReq)
		ok, msg := shouldHaveGrpcErrorStatus(err, "password too weak")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid fields", func(t *testing.T) {
		resp, err := s.SignupWithEmail(context.Background(), validNewUserReq)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp == nil {
			t.Error("response must not be nil")
			return
		}

		if len(resp.UserId) < 3 || len(resp.Roles) < 1 {
			t.Errorf("unexpected UserId or roles: %s", resp)
			return
		}
		if resp.PreferredLanguage != "en" || resp.AccountConfirmed {
			t.Errorf("unexpected PreferredLanguage or AccountConfirmed: %s", resp)
			return
		}
		if len(resp.Profiles) != 1 || resp.SelectedProfile == nil {
			t.Errorf("unexpected profiles: %s", resp)
			return
		}
	})

	t.Run("with duplicate user (same email)", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{
			Email:      "test-signup-1@test.com",
			Password:   "SuperSecurePassword123!§$",
			InstanceId: testInstanceID,
		}
		_, err := s.SignupWithEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		// Try to signup again:
		resp, err := s.SignupWithEmail(context.Background(), req)
		if err == nil || resp != nil {
			t.Errorf("should fail, when user exists already, wrong response: %s", resp)
			return
		}
	})
}

func TestCheckRefreshTokenEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_check_refresh_token_1@test.com",
			},
		},
		{
			Account: models.Account{
				Type:          "email",
				AccountID:     "test_check_refresh_token_2@test.com",
				RefreshTokens: []string{"test-token"},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.CheckRefreshToken(context.Background(), nil)
		if err == nil {
			t.Error("should return error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.RefreshTokenRequest{}
		resp, err := s.CheckRefreshToken(context.Background(), req)
		if err == nil {
			t.Error("should return error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with no tokens for the user", func(t *testing.T) {
		req := &api.RefreshTokenRequest{
			RefreshToken: "test-token",
			InstanceId:   testInstanceID,
			UserId:       testUsers[0].ID.Hex(),
		}
		_, err := s.CheckRefreshToken(context.Background(), req)
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("with wrong token for the user", func(t *testing.T) {
		req := &api.RefreshTokenRequest{
			RefreshToken: "wrong-test-token",
			InstanceId:   testInstanceID,
			UserId:       testUsers[1].ID.Hex(),
		}
		_, err := s.CheckRefreshToken(context.Background(), req)
		if err == nil {
			t.Error("should return error")
		}
	})

	t.Run("with token for the user", func(t *testing.T) {
		req := &api.RefreshTokenRequest{
			RefreshToken: "test-token",
			InstanceId:   testInstanceID,
			UserId:       testUsers[1].ID.Hex(),
		}
		_, err := s.CheckRefreshToken(context.Background(), req)
		if err != nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			return
		}

		user, err := getUserByIDFromDB(testInstanceID, testUsers[1].ID.Hex())
		if err != nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			return
		}
		if user.HasRefreshToken("test-token") {
			t.Errorf("refresh token should have been deleted: %s", user.Account.RefreshTokens)
		}
	})
}

func TestTokenRefreshedEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_token_refreshed_1@test.com",
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.TokenRefreshed(context.Background(), nil)
		if err == nil {
			t.Error("should return error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.RefreshTokenRequest{}
		resp, err := s.TokenRefreshed(context.Background(), req)
		if err == nil {
			t.Error("should return error")
		}
		st, ok := status.FromError(err)
		if !ok || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with new token for the user", func(t *testing.T) {
		req := &api.RefreshTokenRequest{
			RefreshToken: "new-test-token",
			InstanceId:   testInstanceID,
			UserId:       testUsers[0].ID.Hex(),
		}
		_, err := s.TokenRefreshed(context.Background(), req)
		if err != nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			return
		}

		user, err := getUserByIDFromDB(testInstanceID, testUsers[0].ID.Hex())
		if err != nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			return
		}
		if !user.HasRefreshToken("new-test-token") {
			t.Errorf("refresh token should have been added: %s", user.Account.RefreshTokens)
		}
	})
}
func TestSwitchProfileEndpoint(t *testing.T) {
	t.Error("test unimplemented")
}
