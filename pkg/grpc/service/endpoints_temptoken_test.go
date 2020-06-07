package service

import (
	"context"
	"testing"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"google.golang.org/grpc/status"
)

func TestGetOrCreateTemptokenEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testTempToken := models.TempToken{
		UserID:     "test_user_id",
		InstanceID: testInstanceID,
		Purpose:    "test_purpose_get_or_create_token",
		Info: map[string]string{
			"key": "test_info",
		},
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err := testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		_, err := s.GetOrCreateTemptoken(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		_, err := s.GetOrCreateTemptoken(context.Background(), &api.TempTokenInfo{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with not existing token", func(t *testing.T) {
		resp, err := s.GetOrCreateTemptoken(context.Background(), &api.TempTokenInfo{
			Purpose:    testTempToken.Purpose,
			UserId:     "otheruserhere",
			InstanceId: testInstanceID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Token) < 1 {
			t.Error("token should be found.")
			return
		}
	})

	t.Run("with existing token", func(t *testing.T) {
		resp, err := s.GetOrCreateTemptoken(context.Background(), &api.TempTokenInfo{
			Purpose:    testTempToken.Purpose,
			UserId:     testTempToken.UserID,
			InstanceId: testInstanceID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Token != testTempToken.Token {
			t.Errorf("unexpected token: %s,  want: %s", resp.Token, testTempToken.Token)
			return
		}
	})
}

func TestGenerateTempTokenEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testTempToken := &api.TempTokenInfo{
		UserId:     "test_user_id",
		InstanceId: testInstanceID,
		Purpose:    "test_purpose",
		Info: map[string]string{
			"key": "test_info",
		},
	}

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.GenerateTempToken(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		resp, err := s.GenerateTempToken(context.Background(), &api.TempTokenInfo{})
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with valid TempToken", func(t *testing.T) {
		resp, err := s.GenerateTempToken(context.Background(), testTempToken)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Token == "" {
			t.Errorf("wrong response: %s", resp)
		}
	})
}

func TestGetTempTokensEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testTempToken := models.TempToken{
		UserID:     "test_user_id",
		InstanceID: testInstanceID,
		Purpose:    "test_purpose_get_tokens",
		Info: map[string]string{
			"key": "test_info",
		},
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err := testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.GetTempTokens(context.Background(), nil)
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		resp, err := s.GetTempTokens(context.Background(), &api.TempTokenInfo{})
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("get by user_id + instace_id", func(t *testing.T) {
		resp, err := s.GetTempTokens(context.Background(), &api.TempTokenInfo{
			UserId:     testTempToken.UserID,
			InstanceId: testTempToken.InstanceID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.TempTokens) < 1 {
			t.Error("token should be found.")
			return
		}
	})

	t.Run("get by user_id + instace_id + type", func(t *testing.T) {
		resp, err := s.GetTempTokens(context.Background(), &api.TempTokenInfo{
			UserId:     testTempToken.UserID,
			InstanceId: testTempToken.InstanceID,
			Purpose:    testTempToken.Purpose,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.TempTokens) != 1 {
			t.Error("exactly one token should be found.")
			return
		}
	})
}

func TestDeleteTempTokenEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testTempToken := models.TempToken{
		UserID:     "test_user_id",
		InstanceID: testInstanceID,
		Purpose:    "test_purpose_delete_token",
		Info: map[string]string{
			"key": "test_info",
		},
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err := testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.DeleteTempToken(context.Background(), nil)
		if err == nil {
			t.Errorf("or response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		resp, err := s.DeleteTempToken(context.Background(), &api.TempToken{})
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with not existing token", func(t *testing.T) {
		resp, err := s.DeleteTempToken(context.Background(), &api.TempToken{
			Token: testTempToken.Token + "1",
		})
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		tt, err := testGlobalDBService.GetTempToken(testTempToken.Token)
		if err != nil || len(tt.Token) < 5 {
			t.Error("token should not be deleted yet")
			return
		}
	})

	t.Run("with existing token", func(t *testing.T) {
		_, err := s.DeleteTempToken(context.Background(), &api.TempToken{
			Token: testTempToken.Token,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		tt, err := testGlobalDBService.GetTempToken(testTempToken.Token)
		if err == nil || len(tt.Token) > 0 {
			t.Error("token should be deleted by now")
			return
		}
	})
}

func TestPurgeUserTempTokensEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenMinimumAgeMin:  time.Second * 1,
			TokenExpiryInterval: time.Second * 2,
		},
	}

	testTempToken := models.TempToken{
		UserID:     "test_user_id",
		InstanceID: testInstanceID,
		Purpose:    "test_purpose_purging",
		Info: map[string]string{
			"key": "test_info",
		},
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	token, err := testGlobalDBService.AddTempToken(testTempToken)
	if err != nil {
		t.Error(err)
		return
	}
	testTempToken.Token = token

	t.Run("without payload", func(t *testing.T) {
		resp, err := s.PurgeUserTempTokens(context.Background(), nil)
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		resp, err := s.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{})
		if err == nil || resp != nil {
			t.Errorf("wrong response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "missing argument" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("with not exisiting wrong instance_id", func(t *testing.T) {
		_, err := s.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{
			InstanceId: testTempToken.InstanceID + "1",
			UserId:     testTempToken.UserID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		tokens, err := testGlobalDBService.GetTempTokenForUser(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tokens) < 1 {
			t.Error("tokens shouldn't be purged yet")
			return
		}
	})

	t.Run("with not exisiting wrong user_id", func(t *testing.T) {
		_, err := s.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{
			InstanceId: testTempToken.InstanceID,
			UserId:     testTempToken.UserID + "1",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		tokens, err := testGlobalDBService.GetTempTokenForUser(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tokens) < 1 {
			t.Error("tokens shouldn't be purged yet")
			return
		}
	})

	t.Run("with exisiting user_id/instance_id combination", func(t *testing.T) {
		resp, err := s.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{
			InstanceId: testTempToken.InstanceID,
			UserId:     testTempToken.UserID,
		})
		if err != nil || resp == nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		tokens, err := testGlobalDBService.GetTempTokenForUser(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tokens) > 0 {
			t.Error("tokens should be all purged")
			return
		}
	})
}
