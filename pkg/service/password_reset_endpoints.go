package service

import (
	"context"
	"log"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/pwhash"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"github.com/influenzanet/user-management-service/pkg/utils"
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
	return nil, status.Error(codes.Unimplemented, "unimplemented: send email")
}

func (s *userManagementServer) GetInfosForPasswordReset(ctx context.Context, req *api.GetInfosForResetPasswordMsg) (*api.UserInfoForPWReset, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	tokenInfos, err := s.ValidateTempToken(req.Token, "password-reset")
	if err != nil {
		log.Printf("GetInfosForPasswordReset: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong token")
	}

	user, err := s.userDBservice.GetUserByID(tokenInfos.InstanceID, tokenInfos.UserID)
	if err != nil {
		log.Printf("GetInfosForPasswordReset: %s", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.UserInfoForPWReset{
		AccountId: user.Account.AccountID,
	}, nil
}

func (s *userManagementServer) ResetPassword(ctx context.Context, req *api.ResetPasswordMsg) (*api.ServiceStatus, error) {
	if req == nil || req.Token == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	tokenInfos, err := s.ValidateTempToken(req.Token, "password-reset")
	if err != nil {
		log.Printf("GetInfosForPasswordReset: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong token")
	}

	if !utils.CheckPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password, err := pwhash.HashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.userDBservice.UpdateUserPassword(tokenInfos.InstanceID, tokenInfos.UserID, password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Printf("user %s initiated password change", tokenInfos.UserID)

	// TODO: initiate email notification for user about password update
	log.Println("TODO: Password Reset - send email")

	// remove all temptokens for password reset:
	if err := s.globalDBService.DeleteAllTempTokenForUser(tokenInfos.InstanceID, tokenInfos.UserID, "password-reset"); err != nil {
		log.Printf("ChangePassword: %s", err.Error())
	}

	// TODO: check password strength
	// TODO: hash password
	// TODO: update passwords
	return &api.ServiceStatus{
		Version: apiVersion,
		Msg:     "password changed",
		Status:  api.ServiceStatus_NORMAL,
	}, status.Error(codes.Unimplemented, "unimplemented: send email about successful password reset")
}
