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
	Name           Name   `bson:"name"`
}

func accountFromAPI(a *user_api.User_Account) Account {
	if a == nil {
		return Account{}
	}
	return Account{
		Type:           a.Type,
		Email:          a.Email,
		EmailConfirmed: a.EmailConfirmed,
		Name:           nameFromAPI(a.Name),
	}
}

// ToAPI converts the object from DB to API format
func (a Account) ToAPI() *user_api.User_Account {
	return &user_api.User_Account{
		Type:           a.Type,
		Email:          a.Email,
		EmailConfirmed: a.EmailConfirmed,
		Name:           a.Name.ToAPI(),
	}
}

// Name holds name properties of a user
type Name struct {
	Gender    string `bson:"gender"`
	Title     string `bson:"title"`
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}

func nameFromAPI(a *user_api.Name) Name {
	if a == nil {
		return Name{}
	}
	return Name{
		Gender:    a.Gender,
		Title:     a.Title,
		FirstName: a.FirstName,
		LastName:  a.LastName,
	}
}

// ToAPI converts the object from DB to API format
func (a Name) ToAPI() *user_api.Name {
	return &user_api.Name{
		Gender:    a.Gender,
		Title:     a.Title,
		FirstName: a.FirstName,
		LastName:  a.LastName,
	}
}
