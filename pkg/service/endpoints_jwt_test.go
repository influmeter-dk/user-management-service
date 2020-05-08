package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/influenzanet/authentication-service/api"
	api_mock "github.com/influenzanet/authentication-service/mocks"
	"github.com/influenzanet/authentication-service/tokens"
)

func TestLoginWithEmail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockUserManagementClient := api_mock.NewMockUserManagementApiClient(mockCtrl)
	clients.userManagement = mockUserManagementClient

	s := authServiceServer{}

	t.Run("Testing login without payload", func(t *testing.T) {
		_, err := s.LoginWithEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid username and/or password")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("Testing login with empty payload", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{}
		mockUserManagementClient.EXPECT().LoginWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, errors.New("invalid username and/or password"))

		_, err := s.LoginWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid username and/or password")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("Testing login with wrong email", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      "wrong@test.com",
			Password:   "dfdfbmdpfbmd",
			InstanceId: testInstanceID,
		}
		mockUserManagementClient.EXPECT().LoginWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, errors.New("invalid username and/or password"))

		_, err := s.LoginWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid username and/or password")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("Testing login with wrong password", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      "test@test.com",
			Password:   "wrongpw",
			InstanceId: testInstanceID,
		}
		mockUserManagementClient.EXPECT().LoginWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, errors.New("invalid username and/or password"))

		_, err := s.LoginWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "invalid username and/or password")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("Testing login with correct email and password", func(t *testing.T) {
		req := &api.LoginWithEmailMsg{
			Email:      "test@test.com",
			Password:   "dfdfbmdpfbmd",
			InstanceId: testInstanceID,
		}

		mockUserManagementClient.EXPECT().LoginWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.UserAuthInfo{
			UserId:           "testid",
			Roles:            []string{"PARTICIPANT", "RESEARCHER"},
			InstanceId:       testInstanceID,
			AccountId:        "test@test.com",
			AccountConfirmed: false,
			Profiles: []*api.Profile{
				{Id: "testprofile_id", Nickname: "test"},
			},
			SelectedProfile:   &api.Profile{Id: "testprofile_id", Nickname: "test"},
			PreferredLanguage: "en",
		}, nil)
		mockUserManagementClient.EXPECT().TokenRefreshed(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)

		resp, err := s.LoginWithEmail(context.Background(), req)
		if err != nil {
			st, _ := status.FromError(err)
			t.Errorf("unexpected error: %s", st.Message())
			return
		}
		if len(resp.AccessToken) < 1 || len(resp.RefreshToken) < 1 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if resp.SelectedProfileId != "testprofile_id" {
			t.Errorf("unexpected selected profile: %s", resp.SelectedProfileId)
			return
		}
		if len(resp.Profiles) != 1 {
			t.Errorf("unexpected number of profiles: %d", len(resp.Profiles))
			return
		}
	})
}

func TestSignup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockUserManagementClient := api_mock.NewMockUserManagementApiClient(mockCtrl)
	clients.userManagement = mockUserManagementClient

	s := authServiceServer{}

	t.Run("Testing signup without payload", func(t *testing.T) {
		_, err := s.SignupWithEmail(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{}
		mockUserManagementClient.EXPECT().SignupWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.InvalidArgument, "missing arguments"))
		_, err := s.SignupWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with too short password", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{
			Email:      "test@test.com",
			Password:   "short",
			InstanceId: testInstanceID,
		}
		mockUserManagementClient.EXPECT().SignupWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.InvalidArgument, "password too weak"))

		_, err := s.SignupWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "password too weak")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with invalid email", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{
			Email:      "test-test.com",
			Password:   "short",
			InstanceId: testInstanceID,
		}
		mockUserManagementClient.EXPECT().SignupWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.InvalidArgument, "email not valid"))

		_, err := s.SignupWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "email not valid")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with existing user", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{
			Email:      "test@test.com",
			Password:   "short",
			InstanceId: testInstanceID,
		}
		mockUserManagementClient.EXPECT().SignupWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.Internal, "user already exists"))

		_, err := s.SignupWithEmail(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "user already exists")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid arguments", func(t *testing.T) {
		req := &api.SignupWithEmailMsg{
			Email:             "test@test.com",
			Password:          "short",
			InstanceId:        testInstanceID,
			PreferredLanguage: "de",
			WantsNewsletter:   false,
		}

		mockUserManagementClient.EXPECT().SignupWithEmail(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.UserAuthInfo{
			UserId:           "testid",
			Roles:            []string{"participant"},
			InstanceId:       testInstanceID,
			AccountId:        "test@test.com",
			AccountConfirmed: false,
			Profiles: []*api.Profile{
				{Id: "testprofile_id", Nickname: "test"},
			},
			SelectedProfile:   &api.Profile{Id: "testprofile_id", Nickname: "test"},
			PreferredLanguage: req.PreferredLanguage,
		}, nil)
		mockUserManagementClient.EXPECT().TokenRefreshed(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)

		resp, err := s.SignupWithEmail(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.AccessToken) < 1 || len(resp.RefreshToken) < 1 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if resp.SelectedProfileId != "testprofile_id" {
			t.Errorf("unexpected selected profile: %s", resp.SelectedProfileId)
			return
		}
		if len(resp.Profiles) != 1 {
			t.Errorf("unexpected number of profiles: %d", len(resp.Profiles))
			return
		}
	})
}

func TestSwitchProfile(t *testing.T) {
	s := authServiceServer{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockUserManagementClient := api_mock.NewMockUserManagementApiClient(mockCtrl)
	clients.userManagement = mockUserManagementClient

	t.Run("without payload", func(t *testing.T) {
		_, err := s.SwitchProfile(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty payload", func(t *testing.T) {
		req := &api.ProfileRequest{}
		mockUserManagementClient.EXPECT().SwitchProfile(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.InvalidArgument, "missing arguments"))
		_, err := s.SwitchProfile(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing arguments")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with not existing profile", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &api.TokenInfos{
				Id:         "userid",
				InstanceId: testInstanceID,
			},
			Profile: &api.Profile{
				Id: "profile_id",
			},
		}
		mockUserManagementClient.EXPECT().SwitchProfile(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, status.Error(codes.InvalidArgument, "profile not found"))
		_, err := s.SwitchProfile(context.Background(), req)
		ok, msg := shouldHaveGrpcErrorStatus(err, "profile not found")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with existing profile", func(t *testing.T) {
		req := &api.ProfileRequest{
			Token: &api.TokenInfos{
				Id:         "userid",
				InstanceId: testInstanceID,
			},
			Profile: &api.Profile{
				Id: "testprofile_id",
			},
		}
		mockUserManagementClient.EXPECT().SwitchProfile(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.UserAuthInfo{
			UserId:           "testid",
			Roles:            []string{"participant"},
			InstanceId:       testInstanceID,
			AccountId:        "test@test.com",
			AccountConfirmed: false,
			Profiles: []*api.Profile{
				{Id: "main", Nickname: "test"},
				{Id: "testprofile_id", Nickname: "test"},
			},
			SelectedProfile:   &api.Profile{Id: "testprofile_id", Nickname: "test"},
			PreferredLanguage: "en",
		}, nil)
		mockUserManagementClient.EXPECT().TokenRefreshed(
			gomock.Any(),
			gomock.Any(),
		).Return(&api.Status{}, nil)
		resp, err := s.SwitchProfile(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.AccessToken) < 1 || len(resp.RefreshToken) < 1 {
			t.Errorf("unexpected response: %s", resp)
			return
		}
		if resp.SelectedProfileId != "testprofile_id" {
			t.Errorf("unexpected selected profile: %s", resp.SelectedProfileId)
			return
		}
		if len(resp.Profiles) != 2 {
			t.Errorf("unexpected number of profiles: %d", len(resp.Profiles))
			return
		}
	})
}

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
