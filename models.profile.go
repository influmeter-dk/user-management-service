package main

import (
	user_api "github.com/influenzanet/api/dist/go/user-management"
)

type Person struct {
	Gender    string `bson:"gender"`
	Title     string `bson:"tite"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
}

func personFromProtobuf(p *user_api.Profile) Person {
	return Person{
		Gender:    "", // TODO
		Title:     "", // TODO
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}

// ToProtobuf converts a person from DB format into the API format
func (p Person) ToProtobuf() *user_api.Profile {
	return &user_api.Profile{
		// TODO: convert attributes like gender, title etc.
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}
