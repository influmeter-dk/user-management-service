package service

import (
	"errors"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
)

func (s *userManagementServer) ValidateTempToken(token string, withPurpose string) (tt *models.TempToken, err error) {
	tokenInfos, err := s.globalDBService.GetTempToken(token)
	if err != nil {
		return nil, errors.New("wrong token")
	}

	if time.Now().Unix() > tokenInfos.Expiration {
		err = s.globalDBService.DeleteTempToken(tokenInfos.Token)
		return nil, err
	}

	if withPurpose != "" && withPurpose != tokenInfos.Purpose {
		return nil, errors.New("wrong token purpose")
	}

	tt = &tokenInfos
	return
}
