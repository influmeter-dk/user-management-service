package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/influenzanet/user-management-service/utils"

	api "github.com/influenzanet/user-management-service/api"
	api_mock "github.com/influenzanet/user-management-service/mocks"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/status"
)

func TestGetUserEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:  "email",
				Email: "get_user_1@test.com",
			},
		},
		User{
			Account: Account{
				Type:  "email",
				Email: "get_user_2@test.com",
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
			Token:  "mck_token",
			UserId: testUsers[0].ID.Hex() + "w",
		}

		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex() + "w",
			InstanceId: testInstanceID,
		}, nil)

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
			Token:  "mck_token",
			UserId: testUsers[1].ID.Hex(),
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
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
			Token:  "mck_token",
			UserId: testUsers[1].ID.Hex(),
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[1].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.GetUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if testUsers[1].Account.Email != resp.Account.Email {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestChangePasswordEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	oldPassword := "SuperSecurePassword123!§$"
	newPassword := "NewSuperSecurePassword123!§$"

	hashedOldPassword, _ := utils.HashPassword(oldPassword)

	// Create Test User
	testUser := User{
		Account: Account{
			Type:     "email",
			Email:    "test-password-change@test.com",
			Password: hashedOldPassword,
		},
		Roles: []string{"PARTICIPANT"},
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
			Token: "mck_token",
			/*Auth: &influenzanet.ParsedToken{
				UserId:     "test-wrong-id",
				Roles:      []string{"PARTICIPANT"},
				InstanceId: testInstanceID,
			},*/
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         "wrong-id",
			InstanceId: testInstanceID,
		}, nil)

		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong old password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token:       "mck_token",
			OldPassword: oldPassword + "wrong",
			NewPassword: newPassword,
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         id,
			InstanceId: testInstanceID,
		}, nil)

		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "invalid user and/or password" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with too weak new password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token:       "mck_token",
			OldPassword: oldPassword,
			NewPassword: "short",
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         id,
			InstanceId: testInstanceID,
		}, nil)

		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "new password too weak" || resp != nil {
			t.Errorf("wrong error: %s", st.Message())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with valid data and new password", func(t *testing.T) {
		req := &api.PasswordChangeMsg{
			Token:       "mck_token",
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         id,
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.ChangePassword(context.Background(), req)
		if err != nil || resp == nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			t.Errorf("or missing response: %s", resp)
		}

		// Check login with new credentials:
		req2 := &api.UserCredentials{
			Email:      testUser.Account.Email,
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

func TestChangeEmailEndpoint(t *testing.T) {
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

func TestUpdateNameEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	// Create Test User
	testUser := User{
		Account: Account{
			Type:  "email",
			Email: "test-name-change@test.com",
			Name: Name{
				Gender:    "Male",
				FirstName: "First",
				LastName:  "Last",
			},
		},
		Roles: []string{"PARTICIPANT"},
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
		resp, err := s.UpdateName(context.Background(), nil)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("without empty fields", func(t *testing.T) {
		req := &api.NameUpdateRequest{}
		resp, err := s.UpdateName(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.NameUpdateRequest{
			Token: "mock-token",
			Name: &api.Name{
				Gender:    "Female",
				FirstName: "First2",
				LastName:  "Last2",
				Title:     "Dr.",
			},
		}

		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         "wrong-id",
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.UpdateName(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "not found" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with valid input", func(t *testing.T) {
		newName := api.Name{
			Gender:    "Female",
			FirstName: "First2",
			LastName:  "Last2",
			Title:     "Dr.",
		}
		req := &api.NameUpdateRequest{
			Token: "mock_token",
			Name:  &newName,
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUser.ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.UpdateName(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.Name == &newName {
			t.Error("name is not updated")
		}
	})
}

func TestDeleteAccountEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	// Create Test User
	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:  "email",
				Email: "delete_user_1@test.com",
			},
		},
		User{
			Account: Account{
				Type:  "email",
				Email: "delete_user_2@test.com",
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
			Token:  "mock_token",
			UserId: testUsers[1].ID.Hex(),
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
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
			Token:  "mock_token",
			UserId: testUsers[0].ID.Hex(),
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
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

func TestUpdateBirthDateEndpoint(t *testing.T) {
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
	// TODO: with own user id - check if values really updated
	t.Error("test not implemented")
}

func TestUpdateChildrenEndpoint(t *testing.T) {
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
	// TODO: with own user id - check if values updated
	t.Error("test not implemented")
}

func TestAddSubprofileEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:  "email",
				Email: "add_subprofile_1@test.com",
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.AddSubprofile(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.SubProfileRequest{}
		resp, err := s.AddSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex() + "w",
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.AddSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with own user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.AddSubprofile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.SubProfiles) < 1 {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestEditSubprofileEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:  "email",
				Email: "edit_subprofile_1@test.com",
			},
			SubProfiles: SubProfiles{
				SubProfile{
					ID:        primitive.NewObjectID(),
					Name:      "Test to Edit",
					BirthYear: 1999,
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.EditSubprofile(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.SubProfileRequest{}
		resp, err := s.EditSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex(),
				Name:      "Testname",
				BirthYear: 1911,
			},
		}

		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex() + "w",
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.EditSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong subprofile id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex() + "1",
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.EditSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "item with given ID not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with own user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex(),
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.EditSubprofile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.SubProfiles) < 1 || resp.SubProfiles[0].Name != "Testname" || resp.SubProfiles[0].BirthYear != 1911 {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestRemoveSubprofileEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuthServiceClient := api_mock.NewMockAuthServiceApiClient(mockCtrl)
	clients.authService = mockAuthServiceClient

	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:  "email",
				Email: "remove_subprofile_1@test.com",
			},
			SubProfiles: SubProfiles{
				SubProfile{
					ID:        primitive.NewObjectID(),
					Name:      "Test to Edit",
					BirthYear: 1999,
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.RemoveSubprofile(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.SubProfileRequest{}
		resp, err := s.RemoveSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id: testUsers[0].SubProfiles[0].ID.Hex(),
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex() + "w",
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.RemoveSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong subprofile id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id: testUsers[0].SubProfiles[0].ID.Hex() + "1",
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.RemoveSubprofile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "item with given ID not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with own user id", func(t *testing.T) {
		req := &api.SubProfileRequest{
			Token: "mock_token",
			SubProfile: &api.SubProfile{
				Id: testUsers[0].SubProfiles[0].ID.Hex(),
			},
		}
		mockAuthServiceClient.EXPECT().ValidateJWT(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.TokenInfos{
			Id:         testUsers[0].ID.Hex(),
			InstanceId: testInstanceID,
		}, nil)
		resp, err := s.RemoveSubprofile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.SubProfiles) > 0 {
			t.Errorf("wrong response: %s", resp)
		}
	})
}
