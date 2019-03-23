package main

import (
	"context"
	"testing"

	influenzanet "github.com/influenzanet/api/dist/go"
	user_api "github.com/influenzanet/api/dist/go/user-management"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/status"
)

func TestGetUserEndpoint(t *testing.T) {
	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:     "email",
				Email:    "get_user_1@test.com",
				Password: hashPassword("13 ckld fg§$5"),
			},
		},
		User{
			Account: Account{
				Type:     "email",
				Email:    "get_user_2@test.com",
				Password: hashPassword("13 ckld fg§$5"),
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
		req := &user_api.UserReference{}
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
		req := &user_api.UserReference{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex() + "w",
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
		req := &user_api.UserReference{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[1].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[0].ID.Hex(),
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
		req := &user_api.UserReference{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[1].ID.Hex(),
				InstanceId: testInstanceID,
			},
			UserId: testUsers[1].ID.Hex(),
		}
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
	s := userManagementServer{}

	oldPassword := "SuperSecurePassword123!§$"
	newPassword := "NewSuperSecurePassword123!§$"

	// Create Test User
	testUser := User{
		Account: Account{
			Type:     "email",
			Email:    "test-password-change@test.com",
			Password: hashPassword(oldPassword),
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
		req := &user_api.PasswordChangeMsg{}
		resp, err := s.ChangePassword(context.Background(), req)
		st, ok := status.FromError(err)
		if !ok || st == nil || st.Message() != "missing argument" || resp != nil {
			t.Errorf("wrong error: %s", err.Error())
			t.Errorf("or response: %s", resp)
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
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

	t.Run("with wrong old password", func(t *testing.T) {
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

	t.Run("with too weak new password", func(t *testing.T) {
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

	t.Run("with valid data and new password", func(t *testing.T) {
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
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with wrong email format
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestUpdateNameEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestUpdateBirthDateEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id - check if values really updated
	t.Error("test not implemented")
}

func TestUpdateChildrenEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id - check if values updated
	t.Error("test not implemented")
}

/* TODO: remove
func TestUpdateProfileEndpoint(t *testing.T) {
	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:     "email",
				Email:    "update_profile_1@test.com",
				Password: hashPassword("13sd ckld fg§$5"),
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.UpdateProfile(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &user_api.ProfileRequest{}
		resp, err := s.UpdateProfile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with wrong user id", func(t *testing.T) {
		req := &user_api.ProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex() + "w",
				InstanceId: testInstanceID,
			},
			Profile: &user_api.Profile{
				Gender:    "test",
				Title:     "none",
				FirstName: "First",
				LastName:  "Last",
			},
		}
		resp, err := s.UpdateProfile(context.Background(), req)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "not found" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with own user id", func(t *testing.T) {
		newProfile := &user_api.Profile{
			Gender:    "test",
			Title:     "none",
			FirstName: "First",
			LastName:  "Last",
		}
		req := &user_api.ProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			Profile: newProfile,
		}
		resp, err := s.UpdateProfile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Profile.Gender != newProfile.Gender || resp.Profile.LastName != newProfile.LastName {
			t.Errorf("wrong response: %s", resp)
		}
	})
}
*/

func TestAddSubprofileEndpoint(t *testing.T) {
	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:     "email",
				Email:    "add_subprofile_1@test.com",
				Password: hashPassword("54sd ckld fg§pe5"),
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
		req := &user_api.SubProfileRequest{}
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
		req := &user_api.SubProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex() + "w",
				InstanceId: testInstanceID,
			},
			SubProfile: &user_api.SubProfile{
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
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
		req := &user_api.SubProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			SubProfile: &user_api.SubProfile{
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
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
	s := userManagementServer{}

	testUsers, err := addTestUsers([]User{
		User{
			Account: Account{
				Type:     "email",
				Email:    "edit_subprofile_1@test.com",
				Password: hashPassword("54sd ckld fg§pe5"),
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
		req := &user_api.SubProfileRequest{}
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
		req := &user_api.SubProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex() + "w",
				InstanceId: testInstanceID,
			},
			SubProfile: &user_api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex(),
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
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
		req := &user_api.SubProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			SubProfile: &user_api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex() + "1",
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
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
		req := &user_api.SubProfileRequest{
			Auth: &influenzanet.ParsedToken{
				UserId:     testUsers[0].ID.Hex(),
				InstanceId: testInstanceID,
			},
			SubProfile: &user_api.SubProfile{
				Id:        testUsers[0].SubProfiles[0].ID.Hex(),
				Name:      "Testname",
				BirthYear: 1911,
			},
		}
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
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}
