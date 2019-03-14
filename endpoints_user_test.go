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
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
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

func TestSetProfileEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestAddSubprofileEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestEditSubprofileEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}

func TestRemoveSubprofileEndpoint(t *testing.T) {
	// s := userManagementServer{}

	// TODO: without payload
	// TODO: with empty payload
	// TODO: with other user id
	// TODO: with own user id
	t.Error("test not implemented")
}
