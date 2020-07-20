package service

import (
	"context"
	"strings"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"github.com/influenzanet/user-management-service/pkg/utils"
	"google.golang.org/grpc/codes"

	api_types "github.com/influenzanet/go-utils/pkg/api_types"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) ValidateJWT(ctx context.Context, req *api.JWTRequest) (*api_types.TokenInfos, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	// Parse and validate token
	parsedToken, ok, err := tokens.ValidateToken(req.Token)
	if err != nil || !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	return &api_types.TokenInfos{
		Id:               parsedToken.ID,
		InstanceId:       parsedToken.InstanceID,
		IssuedAt:         parsedToken.IssuedAt,
		AccountConfirmed: parsedToken.AccountConfirmed,
		Payload:          parsedToken.Payload,
		ProfilId:         parsedToken.ProfileID,
		TempToken:        parsedToken.TempTokenInfos.ToAPI(),
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
	newToken, err := tokens.GenerateNewToken(parsedToken.ID, user.Account.AccountConfirmedAt > 0, parsedToken.ProfileID, roles, parsedToken.InstanceID, s.JWT.TokenExpiryInterval, username, nil, parsedToken.OtherProfileIDs)
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

func (s *userManagementServer) RevokeAllRefreshTokens(ctx context.Context, req *api.RevokeRefreshTokensReq) (*api.ServiceStatus, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}

	user, err := s.userDBservice.GetUserByID(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}
	user.Account.RefreshTokens = []string{}

	_, err = s.userDBservice.UpdateUser(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "refresh tokens revoked",
		Version: apiVersion,
	}, nil
}
