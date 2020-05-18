package service

import (
	"context"
	"testing"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestInitiatePasswordResetEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_pwreset_init@test.com",
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_pwreset_init@test.com",
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
		_, err := s.InitiatePasswordReset(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		_, err := s.InitiatePasswordReset(context.Background(), &api.InitiateResetPasswordMsg{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong account id", func(t *testing.T) {
		_, err := s.InitiatePasswordReset(context.Background(), &api.InitiateResetPasswordMsg{
			AccountId: "wrong@test.test",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid account id")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid account id", func(t *testing.T) {
		_, err := s.InitiatePasswordReset(context.Background(), &api.InitiateResetPasswordMsg{
			InstanceId: testInstanceID,
			AccountId:  testUsers[0].Account.AccountID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})
}

func TestGetInfosForPasswordResetEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_pwreset_infos@test.com",
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_pwreset_infos@test.com",
					ConfirmedAt: time.Now().Unix(),
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	testTempTokenOld := models.TempToken{
		UserID:     testUsers[0].ID.Hex(),
		InstanceID: testInstanceID,
		Purpose:    "password-reset",
		Expiration: time.Now().Unix() - 10,
	}
	token, err := testGlobalDBService.AddTempToken(testTempTokenOld)
	if err != nil {
		t.Error(err)
		return
	}
	testTempTokenOld.Token = token

	testTempToken := models.TempToken{
		UserID:     testUsers[0].ID.Hex(),
		InstanceID: testInstanceID,
		Purpose:    "password-reset",
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err = testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		_, err := s.GetInfosForPasswordReset(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		_, err := s.GetInfosForPasswordReset(context.Background(), &api.GetInfosForResetPasswordMsg{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong token", func(t *testing.T) {
		_, err := s.GetInfosForPasswordReset(context.Background(), &api.GetInfosForResetPasswordMsg{
			Token: "token",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong token")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with expired token", func(t *testing.T) {
		_, err := s.GetInfosForPasswordReset(context.Background(), &api.GetInfosForResetPasswordMsg{
			Token: testTempTokenOld.Token,
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong token")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid token", func(t *testing.T) {
		resp, err := s.GetInfosForPasswordReset(context.Background(), &api.GetInfosForResetPasswordMsg{
			Token: testTempToken.Token,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.AccountId != testUsers[0].Account.AccountID {
			t.Errorf("unexpected accountID: %s", resp.AccountId)
		}
	})

}

func TestResetPasswordEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testUsers, err := addTestUsers([]models.User{
		{
			Account: models.Account{
				Type:      "email",
				AccountID: "test_for_pwreset@test.com",
			},
			ContactInfos: []models.ContactInfo{
				{
					ID:          primitive.NewObjectID(),
					Type:        "email",
					Email:       "test_for_pwreset@test.com",
					ConfirmedAt: time.Now().Unix(),
				},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to create testusers: %s", err.Error())
		return
	}

	testTempTokenOld := models.TempToken{
		UserID:     testUsers[0].ID.Hex(),
		InstanceID: testInstanceID,
		Purpose:    "password-reset",
		Expiration: time.Now().Unix() - 10,
	}
	token, err := testGlobalDBService.AddTempToken(testTempTokenOld)
	if err != nil {
		t.Error(err)
		return
	}
	testTempTokenOld.Token = token

	testTempToken := models.TempToken{
		UserID:     testUsers[0].ID.Hex(),
		InstanceID: testInstanceID,
		Purpose:    "password-reset",
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err = testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), &api.ResetPasswordMsg{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong token", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), &api.ResetPasswordMsg{
			Token:       "token",
			NewPassword: "tokmefn4n2p3rnp32mne-sd",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong token")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with expired token", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), &api.ResetPasswordMsg{
			Token:       testTempTokenOld.Token,
			NewPassword: "tokmefn4n2p3rnp32mne-sd",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong token")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with weak password", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), &api.ResetPasswordMsg{
			Token:       testTempToken.Token,
			NewPassword: "123",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "password too weak")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		_, err := s.ResetPassword(context.Background(), &api.ResetPasswordMsg{
			Token:       testTempToken.Token,
			NewPassword: "tokmefn4n2p3rnp32mne-sd",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	})
}
