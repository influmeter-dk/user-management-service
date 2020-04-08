package main

import (
	api "github.com/influenzanet/user-management-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile describes personal profile information for a User
type Profile struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	BirthYear          int32              `bson:"birthYear"`
	BirthMonth         int32              `bson:"birthMonth"`
	BirthDay           int32              `bson:"birthDay"`
	BirthDateUpdatedAt int64              `bson:"birthDateUpdatedAt"`

	Children          Children `bson:"children"`
	ChildrenUpdatedAt int64    `bson:"childrenUpdatedAt"`
}

func profileFromAPI(p *api.Profile) Profile {
	return Profile{
		BirthYear:          p.BirthYear,
		BirthMonth:         p.BirthMonth,
		BirthDay:           p.BirthDay,
		BirthDateUpdatedAt: p.BirthDateUpdatedAt,
		Children:           childrenFromAPI(p.Children),
		ChildrenUpdatedAt:  p.ChildrenUpdatedAt,
	}
}

// ToAPI converts a person from DB format into the API format
func (p Profile) ToAPI() *api.Profile {
	return &api.Profile{
		BirthYear:          p.BirthYear,
		BirthMonth:         p.BirthMonth,
		BirthDay:           p.BirthDay,
		BirthDateUpdatedAt: p.BirthDateUpdatedAt,
		Children:           p.Children.ToAPI(),
		ChildrenUpdatedAt:  p.ChildrenUpdatedAt,
	}
}

// Child contains information from a user's child
type Child struct {
	BirthYear int32  `bson:"birthYear"`
	Gender    string `bson:"gender"`
}

// Children is a slice of Child objects
type Children []Child

// ToAPI converts a child object from DB format into the API format
func (o Child) ToAPI() *api.Child {
	return &api.Child{
		BirthYear: o.BirthYear,
		Gender:    o.Gender,
	}
}

func childFromAPI(o *api.Child) Child {
	return Child{
		BirthYear: o.BirthYear,
		Gender:    o.Gender,
	}
}

// ToAPI converts a list of child object from DB to API format
func (children Children) ToAPI() []*api.Child {
	res := make([]*api.Child, len(children))
	for i, v := range children {
		res[i] = v.ToAPI()
	}
	return res
}

func childrenFromAPI(children []*api.Child) []Child {
	res := make([]Child, len(children))
	for i, v := range children {
		res[i] = childFromAPI(v)
	}
	return res
}
