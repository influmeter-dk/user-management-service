package service

import (
	"errors"
	"log"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
)

func (s *userManagementServer) CleanExpiredTemptokens(offset int64) {
	err := s.globalDBService.DeleteTempTokensExpireBefore("", "", time.Now().Unix()-offset)
	if err != nil {
		log.Printf("unexpected error while deleting expired temp tokens: %v", err)
	}
}

func (s *userManagementServer) ValidateTempToken(token string, withPurpose string) (tt *models.TempToken, err error) {
	tokenInfos, err := s.globalDBService.GetTempToken(token)
	if err != nil {
		return nil, errors.New("wrong token")
	}

	if time.Now().Unix() > tokenInfos.Expiration {
		_ = s.globalDBService.DeleteTempToken(tokenInfos.Token)
		return nil, errors.New("token expired")
	}

	if withPurpose != "" && withPurpose != tokenInfos.Purpose {
		return nil, errors.New("wrong token purpose")
	}

	tt = &tokenInfos
	return
}
