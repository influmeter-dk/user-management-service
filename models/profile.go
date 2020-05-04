package models

import (
	api "github.com/influenzanet/user-management-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile describes personal profile information for a User
type Profile struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	Nickname           string             `bson:"nickname,omitempty"`
	ConsentConfirmedAt int64              `bson:"consentConfirmedAt"`
	CreatedAt          int64              `bson:"createdAt"`
	AvatarID           string             `bson:"avatarID,omitempty"`
}

func ProfileFromAPI(p *api.Profile) Profile {
	if p == nil {
		return Profile{}
	}
	dbProf := Profile{
		Nickname:           p.Nickname,
		ConsentConfirmedAt: p.ConsentConfirmedAt,
		CreatedAt:          p.CreatedAt,
		AvatarID:           p.AvatarId,
	}
	if len(p.Id) > 0 {
		_id, _ := primitive.ObjectIDFromHex(p.Id)
		dbProf.ID = _id
	}
	return dbProf
}

// ToAPI converts a person from DB format into the API format
func (p Profile) ToAPI() *api.Profile {
	return &api.Profile{
		Id:                 p.ID.Hex(),
		Nickname:           p.Nickname,
		ConsentConfirmedAt: p.ConsentConfirmedAt,
		CreatedAt:          p.CreatedAt,
		AvatarId:           p.AvatarID,
	}
}
