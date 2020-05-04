package models

import (
	api "github.com/influenzanet/user-management-service/api"
)

// Account holds information about user authentication methods
type Account struct {
	Type               string   `bson:"type"`
	AccountID          string   `bson:"accountID"`
	AccountConfirmedAt int64    `bson:"accountConfirmedAt"`
	Password           string   `bson:"password"`
	RefreshTokens      []string `bson:"refreshTokens"`
}

func AccountFromAPI(a *api.User_Account) Account {
	if a == nil {
		return Account{}
	}
	return Account{
		Type:               a.Type,
		AccountID:          a.AccountId,
		AccountConfirmedAt: a.AccountConfirmedAt,
	}
}

// ToAPI converts the object from DB to API format
func (a Account) ToAPI() *api.User_Account {
	return &api.User_Account{
		Type:               a.Type,
		AccountId:          a.AccountID,
		AccountConfirmedAt: a.AccountConfirmedAt,
	}
}
