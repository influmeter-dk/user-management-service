package service

import (
	"context"
	"strings"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

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
	if tokens.CheckTokenAgeMaturity(parsedToken.StandardClaims.IssuedAt, s.JWT.TokenMinimumAgeMin) {
		return nil, status.Error(codes.Unavailable, "can't renew token so often")
	}

	user, err := s.userDBservice.GetUserByID(parsedToken.InstanceID, parsedToken.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	err = user.RemoveRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "wrong refresh token")
	}
	user.Timestamps.LastTokenRefresh = time.Now().Unix()

	roles := tokens.GetRolesFromPayload(parsedToken.Payload)
	username := tokens.GetUsernameFromPayload(parsedToken.Payload)

	// Generate new access token:
	newToken, err := tokens.GenerateNewToken(parsedToken.ID, parsedToken.ProfileID, roles, parsedToken.InstanceID, s.JWT.TokenExpiryInterval, username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	newRefreshToken, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user.AddRefreshToken(newRefreshToken)

	user, err = s.userDBservice.UpdateUser(parsedToken.InstanceID, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = s.userDBservice.UpdateUser(parsedToken.InstanceID, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.TokenResponse{
		AccessToken:       newToken,
		RefreshToken:      newRefreshToken,
		ExpiresIn:         int32(s.JWT.TokenExpiryInterval / time.Minute),
		SelectedProfileId: parsedToken.ProfileID,
		Profiles:          user.ToAPI().Profiles,
		PreferredLanguage: user.Account.PreferredLanguage,
	}, nil
}
