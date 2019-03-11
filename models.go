package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ObjectInfos struct {
	LastTokenRefresh int64 `bson:"lastTokenRefresh"`
	LastLogin        int64 `bson:"lastLogin"`
	CreatedAt        int64 `bson:"createdAt"`
	UpdatedAt        int64 `bson:"updatedAt"`
}

// User describes the user as saved in the DB
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Email          string             `bson:"email" json:"email"`
	EmailConfirmed bool               `bson:"email_confirmed" json:"email_confirmed"`
	Password       string             `bson:"password" json:"password"`
	Roles          []string           `bson:"roles" json:"roles"`
	ObjectInfos    ObjectInfos        `bson:"objectInfos"`
	PersonInfos    Person             `bson:"personInfos"`
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
