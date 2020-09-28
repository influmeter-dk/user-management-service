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

func (s *userManagementServer) ValidateTempToken(token string, purposes []string) (tt *models.TempToken, err error) {
	tokenInfos, err := s.globalDBService.GetTempToken(token)
	if err != nil {
		return nil, errors.New("wrong token")
	}

	if time.Now().Unix() > tokenInfos.Expiration {
		_ = s.globalDBService.DeleteTempToken(tokenInfos.Token)
		return &tokenInfos, errors.New("token expired")
	}

	if len(purposes) > 0 {
		found := false
		for _, p := range purposes {
			if p == tokenInfos.Purpose {
				found = true
				break
			}
		}
		if !found {
			return &tokenInfos, errors.New("wrong token purpose")
		}
	}
	tt = &tokenInfos
	return
}
