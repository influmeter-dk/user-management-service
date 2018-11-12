package main

import "github.com/mongodb/mongo-go-driver/bson/objectid"

// TODO: use this file to define data structs and models used across the main package
type User struct {
	ID             objectid.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email          string            `bson:"email" json:"email"`
	EmailConfirmed bool              `bson:"email_confirmed" json:"email_confirmed"`
	Password       string            `bson:"password" json:"password"`
	Roles          []string          `bson:"roles" json:"roles"`
}
