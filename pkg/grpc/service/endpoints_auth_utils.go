package service

import (
	"context"
	"log"
	"time"

	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) generateAndSendVerificationCode(instanceID string, user models.User) error {
	vc, err := tokens.GenerateVerificationCode(6)
	if err != nil {
		log.Printf("unexpected error while generating verification code: %v", err)
		return status.Error(codes.Internal, "error while generating verification code")
	}

	user.Account.VerificationCode = models.VerificationCode{
		Code:      vc,
		ExpiresAt: time.Now().Unix() + verificationCodeLifetime,
	}
	user, err = s.userDBservice.UpdateUser(instanceID, user)
	if err != nil {
		log.Printf("generateAndSendVerificationCode: unexpected error when saving user -> %v", err)
		return status.Error(codes.Internal, "user couldn't be updated")
	}

	// ---> Trigger message sending
	func(instanceID string, accountID string, code string, preferredLang string) {
		_, err = s.clients.MessagingService.SendInstantEmail(context.TODO(), &messageAPI.SendEmailReq{
			InstanceId:  instanceID,
			To:          []string{accountID},
			MessageType: "verification-code",
			ContentInfos: map[string]string{
				"verificationCode": code,
			},
			PreferredLanguage: preferredLang,
		})
		if err != nil {
			log.Printf("SendVerificationCode: %s", err.Error())
		}
	}(instanceID, user.Account.AccountID, vc, user.Account.PreferredLanguage)
	return nil
}
