package main

import (
	"errors"

	user_api "github.com/influenzanet/api/dist/go/user-management"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectInfos describes metadata for the User
type ObjectInfos struct {
	LastTokenRefresh int64 `bson:"lastTokenRefresh"`
	LastLogin        int64 `bson:"lastLogin"`
	CreatedAt        int64 `bson:"createdAt"`
	UpdatedAt        int64 `bson:"updatedAt"`
}

// ToAPI converts the object from DB to API format
func (o ObjectInfos) ToAPI() *user_api.User_Infos {
	return &user_api.User_Infos{
		LastTokenRefresh: o.LastTokenRefresh,
		LastLogin:        o.LastLogin,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}

// User describes the user as saved in the DB
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Account     Account            `bson:"account"`
	Roles       []string           `bson:"roles" json:"roles"`
	ObjectInfos ObjectInfos        `bson:"objectInfos"`
	Profile     Profile            `bson:"profile"`
	SubProfiles []SubProfile       `bson:"subProfiles"` // earlier referred as 'household member'
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

// AddSubProfile generates unique ID and adds sub-profile to the user's array
func (u *User) AddSubProfile(sp SubProfile) {
	sp.ID = primitive.NewObjectID()
	u.SubProfiles = append(u.SubProfiles, sp)
}

// UpdateSubProfile finds and replaces sub-profile in the user's array
func (u *User) UpdateSubProfile(sp SubProfile) error {
	for i, cSP := range u.SubProfiles {
		if cSP.ID == sp.ID {
			u.SubProfiles[i] = sp
			return nil
		}
	}
	return errors.New("item with given ID not found")
}

// RemoveSubProfile finds and removes sub-profile from the user's array
func (u *User) RemoveSubProfile(id string) error {
	for i, cSP := range u.SubProfiles {
		if cSP.ID.Hex() == id {
			u.SubProfiles = append(u.SubProfiles[:i], u.SubProfiles[i+1:]...)
			return nil
		}
	}
	return errors.New("item with given ID not found")
}
