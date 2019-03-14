package main

import (
	user_api "github.com/influenzanet/api/dist/go/user-management"
)

// Profile describes personal profile information for a User
type Profile struct {
	Gender    string `bson:"gender"`
	Title     string `bson:"title"`
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}

func profileFromAPI(p *user_api.Profile) Profile {
	return Profile{
		Gender:    p.Gender,
		Title:     p.Title,
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}

// ToAPI converts a person from DB format into the API format
func (p Profile) ToAPI() *user_api.Profile {
	return &user_api.Profile{
		Gender:    p.Gender,
		Title:     p.Title,
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}
