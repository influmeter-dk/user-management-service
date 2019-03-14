package main

import (
	user_api "github.com/influenzanet/api/dist/go/user-management"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubProfile defines other persons the user can add reports about
type SubProfile struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	BirthYear int32              `bson:"birthYear"`
}

// SubProfiles is a slice of SubProfile objects
type SubProfiles []SubProfile

func subProfileFromAPI(a *user_api.SubProfile) SubProfile {
	if a == nil {
		return SubProfile{}
	}
	_id, _ := primitive.ObjectIDFromHex(a.Id)
	return SubProfile{
		ID:        _id,
		Name:      a.Name,
		BirthYear: a.BirthYear,
	}
}

// ToAPI converts the object from DB to API format
func (sp SubProfile) ToAPI() *user_api.SubProfile {
	return &user_api.SubProfile{
		Id:        sp.ID.Hex(),
		Name:      sp.Name,
		BirthYear: sp.BirthYear,
	}
}

// ToAPI converts a list of object from DB to API format
func (sps SubProfiles) ToAPI() []*user_api.SubProfile {
	res := make([]*user_api.SubProfile, len(sps))
	for i, v := range sps {
		res[i] = v.ToAPI()
	}
	return res
}
