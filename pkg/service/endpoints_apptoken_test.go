package service

import (
	"context"
	"testing"

	"github.com/influenzanet/authentication-service/api"
	"github.com/influenzanet/authentication-service/models"
	"google.golang.org/grpc/status"
)

func TestValidateAppTokenEndpoint(t *testing.T) {
	s := authServiceServer{}

	appToken := models.AppToken{
		AppName:   "testapp",
		Instances: []string{testInstanceID},
		Tokens:    []string{"test1", "test2"},
	}
	ctx, cancel := getContext()
	defer cancel()

	_, err := collectionAppToken().InsertOne(ctx, appToken)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
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
