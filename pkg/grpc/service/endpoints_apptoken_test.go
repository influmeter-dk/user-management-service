package service

import (
	"context"
	"testing"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"google.golang.org/grpc/status"
)

func TestValidateAppTokenEndpoint(t *testing.T) {
	s := userManagementServer{
		userDBservice:   testUserDBService,
		globalDBService: testGlobalDBService,
		JWT: models.JWTConfig{
			TokenExpiryInterval: time.Second * 2,
		},
	}

	appToken := models.AppToken{
		AppName:   "testapp",
		Instances: []string{testInstanceID},
		Tokens:    []string{"test1", "test2"},
	}
	err := testGlobalDBService.AddAppToken(appToken)
	if err != nil {
		t.Errorf("unexpected error when creating app token: %s", err.Error())
		return
	}

	t.Run("Without payload", func(t *testing.T) {
		resp, err := s.ValidateAppToken(context.Background(), nil)
		if err == nil {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "invalid app token" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("With empty request", func(t *testing.T) {
		resp, err := s.ValidateAppToken(context.Background(), &api.AppTokenRequest{})
		if err == nil {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "invalid app token" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("With wrong token", func(t *testing.T) {
		resp, err := s.ValidateAppToken(context.Background(), &api.AppTokenRequest{
			Token: "wrong",
		})
		if err == nil {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if status.Convert(err).Message() != "invalid app token" {
			t.Errorf("wrong error: %s", err.Error())
		}
	})

	t.Run("Validate existing app token", func(t *testing.T) {
		resp, err := s.ValidateAppToken(context.Background(), &api.AppTokenRequest{
			Token: "test2",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Instances) < 1 || resp.Instances[0] != testInstanceID {
			t.Error("wrong response")
		}
	})
}
