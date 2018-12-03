package main

import "github.com/mongodb/mongo-go-driver/bson/objectid"

// TODO: use this file to define data structs and models used across the main package
type User struct {
	ID             objectid.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email          string            `bson:"email" json:"email"`
	EmailConfirmed bool              `bson:"email_confirmed" json:"email_confirmed"`
	Password       string            `bson:"password" json:"password"`
	Roles          []string          `bson:"roles" json:"roles"`
	Role           string            `json:"role"`
}

// HasRole checks whether the user has a specified role
func (u User) HasRole(role string) bool {
	for _, v := range u.Roles {
		if v == role {
			return true
		}
	}
	return false
}

type UserLoginResponse struct {
	ID   string `json:"user_id"`
	Role string `json:"role"`
}
