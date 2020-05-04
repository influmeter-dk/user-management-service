package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/influenzanet/user-management-service/models"
	"github.com/influenzanet/user-management-service/utils"

	api "github.com/influenzanet/user-management-service/api"
	api_mock "github.com/influenzanet/user-management-service/mocks"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/status"
)

func TestGetUserEndpoint(t *testing.T) {
	s := userManagementServer{}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "get_user_1@test.com",
			},
		},
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "get_user_2@test.com",
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.GetUser(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.UserReference{}
		resp, err := s.GetUser(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.UserReference{
			Token: &api.TokenInfos{
				Id:         testUsers[0].ID.Hex() + "w",
				InstanceId: testInstanceID,
			},
			UserId: testUsers[0].ID.Hex() + "w",
		}

		resp, err := s.GetUser(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with other user id", func(t *testing.T) {
		req := &api.UserReference{
			Token: &api.TokenInfos{
				Id:         testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[1].ID.Hex(),
		}

		resp, err := s.GetUser(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not authorized" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with own user id", func(t *testing.T) {
		req := &api.UserReference{
			Token: &api.TokenInfos{
				Id:         testUsers[1].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[1].ID.Hex(),
		}

		resp, err := s.GetUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if testUsers[1].Account.AccountID != resp.Account.AccountId {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestChangePasswordEndpoint(t *testing.T) {
	s := userManagementServer{}

	oldPassword := "SuperSecurePassword123!§$"
	newPassword := "NewSuperSecurePassword123!§$"

	hashedOldPassword, _ := utils.HashPassword(oldPassword)

	// Create Test User
	testUser := models.User{
		Account: models.Account{
			Type:      "email",
			AccountID: "test-password-change@test.com",
			Password:  hashedOldPassword,
		},
		Roles: []string{"PARTICIPANT"},
		Profiles: []models.Profile{
			{ID: primitive.NewObjectID()},
		},
	}

	id, err := addUserToDB(testInstanceID, testUser)
	if err != nil {
		t.Errorf("error creating users for testing pw change")
		return
	}
	testUser.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Errorf("error converting id")
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.ChangePassword(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("without auth fields", func(t *testing.T) {
		req := &api.PasswordChangeMsg{}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token: &api.TokenInfos{
				Id:         "wrong-id",
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

	t.Run("with wrong old password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token: &api.TokenInfos{
				Id:         id,
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

	t.Run("with too weak new password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token: &api.TokenInfos{
				Id:         id,
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

	t.Run("with valid data and new password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token: &api.TokenInfos{
				Id:         id,
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
		req2 := &api.UserCredentials{
			Email:      testUser.Account.AccountID,
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

func TestChangeAccountIDEmailEndpoint(t *testing.T) {
	/*
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
		clients.authService = mockAuthServiceClient
	*/
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with wrong email format
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestDeleteAccountEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	// Create Test User
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "delete_user_1@test.com",
			},
		},
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "delete_user_2@test.com",
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.DeleteAccount(context.Background(), nil)
		if err == nil {
			t.Error("should return error")
			return
		}
		if status.Convert(err).Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.UserReference{}
		resp, err := s.DeleteAccount(context.Background(), req)
		if err == nil {
			t.Error("should return error")
			return
		}
		if status.Convert(err).Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with other user", func(t *testing.T) {
		req := &api.UserReference{
			Token: &api.TokenInfos{
				Id:         testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[1].ID.Hex(),
		}

		resp, err := s.DeleteAccount(context.Background(), req)
		if err == nil {
			t.Error("should return error")
			return
		}
		if status.Convert(err).Message() != "not authorized" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with same user", func(t *testing.T) {
		req := &api.UserReference{
			Token: &api.TokenInfos{
				Id:         testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[0].ID.Hex(),
		}

		mockAuthServiceClient.EXPECT().PurgeUserTempTokens(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)

		_, err := s.DeleteAccount(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		_, err = getUserByIDFromDB(testInstanceID, testUsers[0].ID.Hex())
		if err == nil {
			t.Error("user should not exist")
		}
	})
}

func TestChangePreferredLanguageEndpoint(t *testing.T) {
	t.Error("test unimplemented")
}

func TestSaveProfileEndpoint(t *testing.T) {
	// TODO: use profiles
	t.Error("test unimplemented")
}

func TestRemoveProfileEndpoint(t *testing.T) {
	// TODO: use profiles
	// create new
	// update profile
	t.Error("test unimplemented")
}

func TestUpdateContactPreferencesEndpoint(t *testing.T) {
	// TODO: use profiles
	// create new
	// update profile
	t.Error("test unimplemented")
}

func TestAddEmailEndpoint(t *testing.T) {
	// TODO: use profiles
	// create new
	// update profile
	t.Error("test unimplemented")
}

func TestRemoveEmailEndpoint(t *testing.T) {
	// TODO: use profiles
	// create new
	// update profile
	t.Error("test unimplemented")
}
