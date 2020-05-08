package service

/*
import (
	"context"
	"testing"
	"time"

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
		req2 := &api.LoginWithEmailMsg{
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
	s := userManagementServer{}

	oldEmailContantID := primitive.NewObjectID()
	// Create Test User
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:               "email",
				AccountID:          "change_account_id_0@test.com",
				AccountConfirmedAt: 1231239192,
			},
		},
		{
			Account: models.Account{
				Type:               "email",
				AccountID:          "change_account_id_1@test.com",
				AccountConfirmedAt: 1231239192,
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          oldEmailContantID,
					Type:        "email",
					Email:       "change_account_id_1@test.com",
					ConfirmedAt: 1231239192,
				},
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "change_account_id_1_new@test.com",
					ConfirmedAt: 1231239192,
				},
			},
			ContactPreferences: models.ContactPreferences{
				SendNewsletterTo: []string{oldEmailContantID.Hex()},
			},
		},
		{
			Account: models.Account{
				Type:               "email",
				AccountID:          "change_account_id_2@test.com",
				AccountConfirmedAt: 0,
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "change_account_id_2@test.com",
					ConfirmedAt: 0,
				},
			},
		},
		{
			Account: models.Account{
				Type:               "email",
				AccountID:          "change_account_id_3@test.com",
				AccountConfirmedAt: 123123123,
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "change_account_id_3@test.com",
					ConfirmedAt: 123123123,
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.ChangeAccountIDEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.EmailChangeMsg{}
		_, err := s.ChangeAccountIDEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("try to update to an already existing email", func(t *testing.T) {
		req := &api.EmailChangeMsg{
			Token: &api.TokenInfos{
				Id:         testUsers[1].ID.Hex(),
				InstanceId: testInstanceID,
			},
			NewEmail: testUsers[0].Account.AccountID,
		}
		_, err := s.ChangeAccountIDEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "already in use")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("for confirmed new email", func(t *testing.T) {
		req := &api.EmailChangeMsg{
			Token: &api.TokenInfos{
				Id:         testUsers[1].ID.Hex(),
				InstanceId: testInstanceID,
			},
			NewEmail: testUsers[1].ContactInfos[1].Email,
		}

		resp, err := s.ChangeAccountIDEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.AccountId != testUsers[1].ContactInfos[1].Email {
			t.Errorf("unexpected accountID: %s", resp.Account.AccountId)
			return
		}
		if resp.Account.AccountConfirmedAt <= 0 {
			t.Errorf("unexpected AccountConfirmedAt: %d", resp.Account.AccountConfirmedAt)
			return
		}
		if resp.ContactPreferences.SendNewsletterTo[0] != testUsers[1].ContactInfos[1].ID.Hex() {
			t.Errorf("unexpected contactPreferences: %s", resp)
			return
		}
	})

	t.Run("for not confirmed old email", func(t *testing.T) {
		req := &api.EmailChangeMsg{
			Token: &api.TokenInfos{
				Id:         testUsers[2].ID.Hex(),
				InstanceId: testInstanceID,
			},
			NewEmail: "newemail@test.com",
		}
		resp, err := s.ChangeAccountIDEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.AccountId != req.NewEmail {
			t.Errorf("unexpected accountID: %s", resp.Account.AccountId)
			return
		}
		if resp.Account.AccountConfirmedAt > 0 {
			t.Errorf("unexpected AccountConfirmedAt: %d", resp.Account.AccountConfirmedAt)
			return
		}
	})

	t.Run("for confirmed old email", func(t *testing.T) {
		req := &api.EmailChangeMsg{
			Token: &api.TokenInfos{
				Id:         testUsers[3].ID.Hex(),
				InstanceId: testInstanceID,
			},
			NewEmail: "newemail2@test.com",
		}
		resp, err := s.ChangeAccountIDEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.AccountId != req.NewEmail {
			t.Errorf("unexpected accountID: %s", resp.Account.AccountId)
			return
		}
		if resp.Account.AccountConfirmedAt > 0 {
			t.Errorf("unexpected AccountConfirmedAt: %d", resp.Account.AccountConfirmedAt)
			return
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
		_, err := s.DeleteAccount(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
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
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:              "email",
				AccountID:         "test_for_change_preferred_lang@test.com",
				PreferredLanguage: "de",
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.ChangePreferredLanguage(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.LanguageChangeMsg{}
		_, err := s.ChangePreferredLanguage(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}

	t.Run("with normal payload", func(t *testing.T) {
		req := &api.LanguageChangeMsg{
			Token:        &token,
			LanguageCode: "fr",
		}
		resp, err := s.ChangePreferredLanguage(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Account.PreferredLanguage != "fr" {
			t.Errorf("unexpected language code: %s", resp.Account.PreferredLanguage)
		}
	})
}

func TestSaveProfileEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_save_profile@test.com",
			},
			Profiles: []models.Profile{
				{
					ID:       primitive.NewObjectID(),
					Nickname: "main",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.SaveProfile(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}

	t.Run("with add profile", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &token,
			Profile: &api.Profile{
				Nickname: "new test",
			},
		}
		resp, err := s.SaveProfile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Profiles) != 2 || resp.Profiles[1].Nickname != "new test" {
			t.Errorf("unexpected response code: %s", resp)
		}
	})

	t.Run("with update profile", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &token,
			Profile: &api.Profile{
				Id:       testUsers[0].Profiles[0].ID.Hex(),
				Nickname: "renamed",
			},
		}
		resp, err := s.SaveProfile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Profiles) != 2 || resp.Profiles[0].Nickname != "renamed" {
			t.Errorf("unexpected response code: %s", resp)
		}
	})
}

func TestRemoveProfileEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_remove_profile@test.com",
			},
			Profiles: []models.Profile{
				{
					ID:       primitive.NewObjectID(),
					Nickname: "main",
				},
				{
					ID:       primitive.NewObjectID(),
					Nickname: "new1",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.RemoveProfile(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.ProfileRequest{}
		_, err := s.RemoveProfile(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}
	t.Run("with wrong id", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &token,
			Profile: &api.Profile{
				Id: "wrong id",
			},
		}
		_, err := s.RemoveProfile(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "profile with given ID not found")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with correct id", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &token,
			Profile: &api.Profile{
				Id: testUsers[0].Profiles[0].ID.Hex(),
			},
		}
		resp, err := s.RemoveProfile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Profiles) != 1 || resp.Profiles[0].Nickname == "main" {
			t.Errorf("wrong response: %s", resp)
		}
	})

	t.Run("last one", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &token,
			Profile: &api.Profile{
				Id: testUsers[0].Profiles[1].ID.Hex(),
			},
		}
		_, err := s.RemoveProfile(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "can't delete last profile")
		if !ok {
			t.Error(msg)
		}
	})
}

func TestUpdateContactPreferencesEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_update_contact_prefs@test.com",
			},
			ContactPreferences: models.ContactPreferences{
				SubscribedToNewsletter: true,
				SendNewsletterTo:       []string{"addr_id_1", "addr_id_2"},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.UpdateContactPreferences(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.ContactPreferencesMsg{}
		_, err := s.UpdateContactPreferences(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}

	t.Run("update address list and subscription", func(t *testing.T) {
		req := &api.ContactPreferencesMsg{
			Token: &token,
			ContactPreferences: &api.ContactPreferences{
				SubscribedToNewsletter: false,
				SendNewsletterTo:       []string{"only_here_id"},
			},
		}
		resp, err := s.UpdateContactPreferences(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.ContactPreferences.SendNewsletterTo) != 1 || resp.ContactPreferences.SubscribedToNewsletter {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestAddEmailEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_add_email@test.com",
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_add_email@test.com",
					ConfirmedAt: time.Now().Unix(),
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.AddEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.ContactInfoMsg{}
		_, err := s.AddEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}

	t.Run("with wrong type", func(t *testing.T) {
		req := &api.ContactInfoMsg{
			Token: &token,
			ContactInfo: &api.ContactInfo{
				Type:    "phone",
				Address: &api.ContactInfo_Email{Email: "new_email@test.com"},
			},
		}
		_, err := s.AddEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong contact type")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("add email", func(t *testing.T) {
		req := &api.ContactInfoMsg{
			Token: &token,
			ContactInfo: &api.ContactInfo{
				Type:    "email",
				Address: &api.ContactInfo_Email{Email: "new_email@test.com"},
			},
		}
		resp, err := s.AddEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.ContactInfos) != 2 || len(resp.ContactInfos[1].Id) < 1 {
			t.Errorf("number of contacts: %d", len(resp.ContactInfos))
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestRemoveEmailEndpoint(t *testing.T) {
	s := userManagementServer{}
	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_remove_email@test.com",
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_remove_email@test.com",
					ConfirmedAt: time.Now().Unix(),
				},
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_remove_email1@test.com",
					ConfirmedAt: time.Now().Unix(),
				}, {
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_remove_email2@test.com",
					ConfirmedAt: time.Now().Unix(),
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.RemoveEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.ContactInfoMsg{}
		_, err := s.RemoveEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	token := api.TokenInfos{
		Id:         testUsers[0].ID.Hex(),
		InstanceId: testInstanceID,
	}

	t.Run("with wrong id", func(t *testing.T) {
		req := &api.ContactInfoMsg{
			Token: &token,
			ContactInfo: &api.ContactInfo{
				Id: "wrong_id",
			},
		}
		_, err := s.RemoveEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "contact not found")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("try remove main address", func(t *testing.T) {
		req := &api.ContactInfoMsg{
			Token: &token,
			ContactInfo: &api.ContactInfo{
				Id: testUsers[0].ContactInfos[0].ID.Hex(),
			},
		}
		_, err := s.RemoveEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "cannot remove main address")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with existing id", func(t *testing.T) {
		req := &api.ContactInfoMsg{
			Token: &token,
			ContactInfo: &api.ContactInfo{
				Id: testUsers[0].ContactInfos[1].ID.Hex(),
			},
		}
		resp, err := s.RemoveEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.ContactInfos) != 2 || resp.ContactInfos[1].Id != testUsers[0].ContactInfos[2].ID.Hex() {
			t.Errorf("wrong response: %s", resp)
		}
	})
}
*/
