package service

import (
	"context"
	"log"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) InitiatePasswordReset(ctx context.Context, req *api.InitiateResetPasswordMsg) (*api.ServiceStatus, error) {
	if req == nil || req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	instanceID := req.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}

	user, err := s.userDBservice.GetUserByEmail(instanceID, req.AccountId)
	if err != nil {
		log.Printf("InitiatePasswordReset: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid account id")
	}

	// TempToken for contact verification:
	tempTokenInfos := models.TempToken{
		UserID:     user.ID.Hex(),
		InstanceID: instanceID,
		Purpose:    "password-reset",
		Info: map[string]string{
			"email": user.Account.AccountID,
		},
		Expiration: tokens.GetExpirationTime(time.Hour * 24),
	}
	tempToken, err := s.globalDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Printf("TODO: send email for password reset with %s", tempToken)
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
func (s *userManagementServer) GetInfosForPasswordReset(ctx context.Context, req *api.GetInfosForResetPasswordMsg) (*api.UserInfoForPWReset, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
func (s *userManagementServer) ResetPassword(ctx context.Context, req *api.ResetPasswordMsg) (*api.TokenResponse, error) {
	// TODO: check password strength
	// TODO: hash password
	// TODO: update passwords
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
