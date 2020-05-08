package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AppToken is a database entry for a app token
type AppToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AppName   string             `bson:"appName"`
	Tokens    []string           `bson:"tokens"`
	Instances []string           `bson:"instances"`
}
