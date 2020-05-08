package service

import (
	"context"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/status"

	api "github.com/influenzanet/authentication-service/api"
	"github.com/influenzanet/authentication-service/tokens"
)

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*api.Status, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, req *api.LoginWithEmailMsg) (*api.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}
	resp, err := clients.userManagement.LoginWithEmail(context.Background(), req)
	if err != nil {
		log.Printf("error during login with email: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	// generate tokens
	token, err := tokens.GenerateNewToken(resp.UserId, resp.SelectedProfile.Id, resp.Roles, resp.InstanceId, conf.JWT.TokenExpiryInterval, resp.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// submit token to user management
	_, err = clients.userManagement.TokenRefreshed(context.Background(), &api.RefreshTokenRequest{
		InstanceId:   resp.InstanceId,
		RefreshToken: rt,
		UserId:       resp.UserId,
	})
	if err != nil {
		st := status.Convert(err)
		log.Printf("error during signup with email: %s: %s", st.Code(), st.Message())
		return nil, status.Error(codes.Internal, st.Message())
	}

	return &api.TokenResponse{
		AccessToken:       token,
		RefreshToken:      rt,
		ExpiresIn:         int32(conf.JWT.TokenExpiryInterval / time.Minute),
		Profiles:          resp.Profiles,
		SelectedProfileId: resp.SelectedProfile.Id,
		PreferredLanguage: resp.PreferredLanguage,
	}, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, req *api.SignupWithEmailMsg) (*api.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	resp, err := clients.userManagement.SignupWithEmail(context.Background(), req)
	if err != nil {
		st := status.Convert(err)
		log.Printf("error during signup with email: %s: %s", st.Code(), st.Message())
		return nil, status.Error(codes.Internal, st.Message())
	}

	// generate tokens
	token, err := tokens.GenerateNewToken(resp.UserId, resp.SelectedProfile.Id, resp.Roles, resp.InstanceId, conf.JWT.TokenExpiryInterval, resp.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// submit refresh token to user management
	_, err = clients.userManagement.TokenRefreshed(context.Background(), &api.RefreshTokenRequest{
		InstanceId:   resp.InstanceId,
		RefreshToken: rt,
		UserId:       resp.UserId,
	})
	if err != nil {
		st := status.Convert(err)
		log.Printf("error during signup with email: %s: %s", st.Code(), st.Message())
		return nil, status.Error(codes.Internal, st.Message())
	}

	return &api.TokenResponse{
		AccessToken:       token,
		RefreshToken:      rt,
		ExpiresIn:         int32(conf.JWT.TokenExpiryInterval / time.Minute),
		Profiles:          resp.Profiles,
		SelectedProfileId: resp.SelectedProfile.Id,
		PreferredLanguage: resp.PreferredLanguage,
	}, nil
}

func (s *userManagementServer) SwitchProfile(ctx context.Context, req *api.ProfileRequest) (*api.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	resp, err := clients.userManagement.SwitchProfile(context.Background(), req)
	if err != nil {
		log.Printf("error during switching profiles: %s", err.Error())
		return nil, err
	}

	// generate tokens
	token, err := tokens.GenerateNewToken(resp.UserId, resp.SelectedProfile.Id, resp.Roles, resp.InstanceId, conf.JWT.TokenExpiryInterval, resp.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// submit refresh token to user management
	_, err = clients.userManagement.TokenRefreshed(context.Background(), &api.RefreshTokenRequest{
		InstanceId:   resp.InstanceId,
		RefreshToken: rt,
		UserId:       resp.UserId,
	})
	if err != nil {
		st := status.Convert(err)
		log.Printf("error during signup with email: %s: %s", st.Code(), st.Message())
		return nil, status.Error(codes.Internal, st.Message())
	}

	return &api.TokenResponse{
		AccessToken:       token,
		RefreshToken:      rt,
		ExpiresIn:         int32(conf.JWT.TokenExpiryInterval / time.Minute),
		Profiles:          resp.Profiles,
		SelectedProfileId: resp.SelectedProfile.Id,
		PreferredLanguage: resp.PreferredLanguage,
	}, nil
}

func (s *userManagementServer) ValidateJWT(ctx context.Context, req *api.JWTRequest) (*api.TokenInfos, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	// Parse and validate token
	parsedToken, ok, err := tokens.ValidateToken(req.Token)
	if err != nil || !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	return &api.TokenInfos{
		Id:         parsedToken.ID,
		InstanceId: parsedToken.InstanceID,
		IssuedAt:   parsedToken.IssuedAt,
		Payload:    parsedToken.Payload,
		ProfilId:   parsedToken.ProfileID,
	}, nil
}

func (s *userManagementServer) RenewJWT(ctx context.Context, req *api.RefreshJWTRequest) (*api.TokenResponse, error) {
	if req == nil || req.AccessToken == "" || req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}

	// Parse and validate token
	parsedToken, _, err := tokens.ValidateToken(req.AccessToken)
	if err != nil && !strings.Contains(err.Error(), "token is expired by") {
		return nil, status.Error(codes.PermissionDenied, "wrong access token")
	}

	// Check for too frequent requests:
	if tokens.CheckTokenAgeMaturity(parsedToken.StandardClaims.IssuedAt, conf.JWT.TokenMinimumAgeMin) {
		return nil, status.Error(codes.Unavailable, "can't renew token so often")
	}

	// check refresh token from user management
	_, err = clients.userManagement.CheckRefreshToken(context.Background(), &api.RefreshTokenRequest{
		UserId:       parsedToken.ID,
		RefreshToken: req.RefreshToken,
		InstanceId:   parsedToken.InstanceID,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong refresh token") // err
	}

	roles := tokens.GetRolesFromPayload(parsedToken.Payload)
	username := tokens.GetUsernameFromPayload(parsedToken.Payload)

	// Generate new access token:
	newToken, err := tokens.GenerateNewToken(parsedToken.ID, parsedToken.ProfileID, roles, parsedToken.InstanceID, conf.JWT.TokenExpiryInterval, username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	newRefreshToken, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// submit refresh token to user management
	_, err = clients.userManagement.TokenRefreshed(context.Background(), &api.RefreshTokenRequest{
		UserId:       parsedToken.ID,
		InstanceId:   parsedToken.InstanceID,
		RefreshToken: newRefreshToken,
	})
	if err != nil {
		st := status.Convert(err)
		log.Printf("error during token refresh: %s: %s", st.Code(), st.Message())
		return nil, status.Error(codes.Internal, st.Message())
	}

	return &api.TokenResponse{
		AccessToken:  newToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int32(conf.JWT.TokenExpiryInterval / time.Minute),
	}, nil
}
