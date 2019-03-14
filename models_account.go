package main

import (
	user_api "github.com/influenzanet/api/dist/go/user-management"
)

// Account holds information about user authentication methods
type Account struct {
	Type           string `bson:"type"`
	Email          string `bson:"email"`
	Password       string `bson:"password"`
	EmailConfirmed bool   `bson:"emailConfirmed"`
}

func accountFromAPI(a *user_api.User_Account) Account {
	if a == nil {
		return Account{}
	}
	return Account{
		Type:           a.Type,
		Email:          a.Email,
		EmailConfirmed: a.EmailConfirmed,
	}
}

// ToAPI converts the object from DB to API format
func (a Account) ToAPI() *user_api.User_Account {
	return &user_api.User_Account{
		Type:           a.Type,
		Email:          a.Email,
		EmailConfirmed: a.EmailConfirmed,
	}
}
