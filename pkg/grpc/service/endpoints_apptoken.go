package service

import (
	"context"

	"github.com/influenzanet/user-management-service/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) ValidateAppToken(ctx context.Context, req *api.AppTokenRequest) (*api.AppTokenValidation, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid app token")
	}
	tokenInfos, err := s.globalDBService.FindAppToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app token")
	}
	return &api.AppTokenValidation{
		Instances: tokenInfos.Instances,
	}, nil
}
