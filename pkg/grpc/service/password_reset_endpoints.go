package service

import (
	"context"
	"log"
	"time"

	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
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

	// ---> Trigger message sending
	_, err = s.clients.MessagingService.SendInstantEmail(ctx, &messageAPI.SendEmailReq{
		To:          []string{user.Account.AccountID},
		MessageType: "password-reset",
		ContentInfos: map[string]string{
			"token":      tempToken,
			"validUntil": "24", // hours
		},
		PreferredLanguage: user.Account.PreferredLanguage,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// <---
	return &api.ServiceStatus{
		Msg:     "email sending triggered",
		Version: apiVersion,
		Status:  api.ServiceStatus_NORMAL,
	}, nil
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

	user, err := s.userDBservice.GetUserByID(tokenInfos.InstanceID, tokenInfos.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Trigger message sending
	_, err = s.clients.MessagingService.SendInstantEmail(ctx, &messageAPI.SendEmailReq{
		To:                []string{user.Account.AccountID},
		MessageType:       "password-changed",
		PreferredLanguage: user.Account.PreferredLanguage,
	})
	if err != nil {
		log.Printf("ChangePassword: %s", err.Error())
	}
	// ---

	// remove all temptokens for password reset:
	if err := s.globalDBService.DeleteAllTempTokenForUser(tokenInfos.InstanceID, tokenInfos.UserID, "password-reset"); err != nil {
		log.Printf("ChangePassword: %s", err.Error())
	}

	return &api.ServiceStatus{
		Version: apiVersion,
		Msg:     "password changed",
		Status:  api.ServiceStatus_NORMAL,
	}, nil
}
