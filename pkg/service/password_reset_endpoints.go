package service

import (
	"context"

	"github.com/influenzanet/user-management-service/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) InitiatePasswordReset(ctx context.Context, t *api.InitiateResetPasswordMsg) (*api.ServiceStatus, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
func (s *userManagementServer) GetInfosForPasswordReset(ctx context.Context, t *api.GetInfosForResetPasswordMsg) (*api.UserInfoForPWReset, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
func (s *userManagementServer) ResetPassword(ctx context.Context, t *api.ResetPasswordMsg) (*api.TokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
