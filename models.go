package main

import "github.com/mongodb/mongo-go-driver/bson/primitive"

// User describes the user as saved in the DB
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Email          string             `bson:"email" json:"email"`
	EmailConfirmed bool               `bson:"email_confirmed" json:"email_confirmed"`
	Password       string             `bson:"password" json:"password"`
	NewPassword    string             `bson:"-" json:"newPassword"`
	Roles          []string           `bson:"roles" json:"roles"`
	// TODO: add profile with e.g. firstname, lastname etc.
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
