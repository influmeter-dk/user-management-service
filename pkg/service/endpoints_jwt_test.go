package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	api "github.com/influenzanet/authentication-service/api"
	api_mock "github.com/influenzanet/authentication-service/mocks"
	"github.com/influenzanet/authentication-service/tokens"
)

func TestValidateJWT(t *testing.T) {
	conf.JWT.TokenExpiryInterval = time.Second * 2
	conf.JWT.TokenMinimumAgeMin = time.Second * 1

	s := authServiceServer{}

	t.Run("without payload", func(t *testing.T) {
		_, err := s.ValidateJWT(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.JWTRequest{}
		_, err := s.ValidateJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	adminToken, err1 := tokens.GenerateNewToken("test-admin-id", "testprofid", []string{"PARTICIPANT", "ADMIN"}, testInstanceID, conf.JWT.TokenExpiryInterval, "")
	userToken, err2 := tokens.GenerateNewToken("test-user-id", "testprofid", []string{"PARTICIPANT"}, testInstanceID, conf.JWT.TokenExpiryInterval, "")
	if err1 != nil || err2 != nil {
		t.Errorf("unexpected error: %s or %s", err1, err2)
		return
	}

	t.Run("with wrong token", func(t *testing.T) {
		req := &api.JWTRequest{
			Token: adminToken + "x",
		}

		_, err := s.ValidateJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid token")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with normal user token", func(t *testing.T) {
		req := &api.JWTRequest{
			Token: userToken,
		}

		resp, err := s.ValidateJWT(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		roles := tokens.GetRolesFromPayload(resp.Payload)
		if resp == nil || resp.InstanceId != testInstanceID || resp.Id != "test-user-id" || len(roles) != 1 || roles[0] != "PARTICIPANT" {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	t.Run("with admin token", func(t *testing.T) {
		req := &api.JWTRequest{
			Token: adminToken,
		}

		resp, err := s.ValidateJWT(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		roles := tokens.GetRolesFromPayload(resp.Payload)
		if resp == nil || len(roles) < 2 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	if testing.Short() {
		t.Skip("skipping waiting for token test in short mode, since it has to wait for token expiration.")
	}
	time.Sleep(conf.JWT.TokenExpiryInterval + time.Second)

	t.Run("with expired token", func(t *testing.T) {
		req := &api.JWTRequest{
			Token: adminToken,
		}
		_, err := s.ValidateJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid token")
		if !ok {
			t.Error(msg)
		}
	})
}

func TestRenewJWT(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockUserManagementClient := api_mock.NewMockUserManagementApiClient(mockCtrl)
	clients.userManagement = mockUserManagementClient

	conf.JWT.TokenExpiryInterval = time.Second * 2
	conf.JWT.TokenMinimumAgeMin = time.Second * 1

	userToken, err := tokens.GenerateNewToken("test-user-id", "testprofid", []string{"PARTICIPANT"}, testInstanceID, conf.JWT.TokenExpiryInterval, "")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	refreshToken := "TEST-REFRESH-TOKEN-STRING"

	s := authServiceServer{}

	t.Run("Testing token refresh without token", func(t *testing.T) {
		_, err := s.RenewJWT(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty token", func(t *testing.T) {
		req := &api.RefreshJWTRequest{}

		_, err := s.RenewJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong access token", func(t *testing.T) {
		req := &api.RefreshJWTRequest{
			AccessToken:  userToken + "x",
			RefreshToken: refreshToken,
		}
		_, err := s.RenewJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong access token")
		if !ok {
			t.Error(msg)
		}
	})

	// Test eagerly, when min age not reached yet
	t.Run("too eagerly", func(t *testing.T) {
		req := &api.RefreshJWTRequest{
			AccessToken:  userToken,
			RefreshToken: refreshToken,
		}

		_, err := s.RenewJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "can't renew token so often")
		if !ok {
			t.Error(msg)
		}
	})

	if testing.Short() {
		t.Skip("skipping renew token test in short mode, since it has to wait for token expiration.")
	}

	time.Sleep(conf.JWT.TokenMinimumAgeMin)

	t.Run("with wrong refresh token", func(t *testing.T) {
		req := &api.RefreshJWTRequest{
			AccessToken:  userToken,
			RefreshToken: userToken + "x",
		}
		mockUserManagementClient.EXPECT().CheckRefreshToken(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, errors.New("wrong refresh token"))

		_, err := s.RenewJWT(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "wrong refresh token")
		if !ok {
			t.Error(msg)
		}
	})

	// Test renew after min age reached - wait 2 seconds
	t.Run("with normal tokens", func(t *testing.T) {
		req := &api.RefreshJWTRequest{
			AccessToken:  userToken,
			RefreshToken: refreshToken,
		}

		mockUserManagementClient.EXPECT().CheckRefreshToken(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)
		mockUserManagementClient.EXPECT().TokenRefreshed(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)

		resp, err := s.RenewJWT(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp == nil {
			t.Error("response is missing")
			return
		}
		if len(resp.AccessToken) < 10 || len(resp.RefreshToken) < 10 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})

	time.Sleep(conf.JWT.TokenExpiryInterval)

	// Test with expired token
	t.Run("with expired token", func(t *testing.T) {
		req := &api.RefreshJWTRequest{
			AccessToken:  userToken,
			RefreshToken: refreshToken,
		}
		mockUserManagementClient.EXPECT().CheckRefreshToken(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)
		mockUserManagementClient.EXPECT().TokenRefreshed(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)

		resp, err := s.RenewJWT(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp == nil {
			t.Error("response is missing")
			return
		}
		if len(resp.AccessToken) < 10 || len(resp.RefreshToken) < 10 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
	})
}
