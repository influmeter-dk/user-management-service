package models

import (
	"errors"
	"time"

	api "github.com/influenzanet/user-management-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User describes the user as saved in the DB
type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Account            Account            `bson:"account"`
	Roles              []string           `bson:"roles" json:"roles"`
	Timestamps         Timestamps         `bson:"timestamps"`
	Profiles           []Profile          `bson:"profiles"`
	ContactPreferences ContactPreferences `bson:"contactPreferences"`
	ContactInfos       []ContactInfo      `bson:"contactInfos"`
}

// ToAPI converts the object from DB to API format
func (u User) ToAPI() *api.User {
	profiles := make([]*api.Profile, len(u.Profiles))
	for i, p := range u.Profiles {
		profiles[i] = p.ToAPI()
	}
	contactInfos := make([]*api.ContactInfo, len(u.ContactInfos))
	for i, c := range u.ContactInfos {
		contactInfos[i] = c.ToAPI()
	}
	return &api.User{
		Id:                 u.ID.Hex(),
		Account:            u.Account.ToAPI(),
		Roles:              u.Roles,
		Timestamps:         u.Timestamps.ToAPI(),
		Profiles:           profiles,
		ContactPreferences: u.ContactPreferences.ToAPI(),
		ContactInfos:       contactInfos,
	}
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

// Add a new email address
func (u *User) AddNewEmail(addr string, confirmed bool) {
	contactInfo := ContactInfo{
		ID:          primitive.NewObjectID(),
		Type:        "email",
		ConfirmedAt: 0,
		Email:       addr,
	}
	if confirmed {
		contactInfo.ConfirmedAt = time.Now().Unix()
	}
	u.ContactInfos = append(u.ContactInfos, contactInfo)
}

func (u *User) ConfirmContactInfo(id string) error {
	for i, ci := range u.ContactInfos {
		if ci.ID.Hex() == id {
			u.ContactInfos[i].ConfirmedAt = time.Now().Unix()
			return nil
		}
	}
	return errors.New("item with given ID not found")
}

func (u User) FindContactInfo(t string, addr string) (ContactInfo, bool) {
	for _, ci := range u.ContactInfos {
		if t == "email" && ci.Email == addr {
			return ci, true
		} else if t == "phone" && ci.Phone == addr {
			return ci, true
		}
	}
	return ContactInfo{}, false
}

func (u *User) RemoveContactInfo(id string) error {
	for i, ci := range u.ContactInfos {
		if ci.ID.Hex() == id {
			u.ContactInfos = append(u.ContactInfos[:i], u.ContactInfos[i+1:]...)
			return nil
		}
	}
	return errors.New("item with given ID not found")
}

/*
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
*/

// AddRefreshToken add a new refresh token to the user's list
func (u *User) AddRefreshToken(token string) {
	u.Account.RefreshTokens = append(u.Account.RefreshTokens, token)
	if len(u.Account.RefreshTokens) > 10 {
		_, u.Account.RefreshTokens = u.Account.RefreshTokens[0], u.Account.RefreshTokens[1:]
	}
}

// HasRefreshToken checks weather a user has a particular refresh token
func (u *User) HasRefreshToken(token string) bool {
	for _, t := range u.Account.RefreshTokens {
		if t == token {
			return true
		}
	}
	return false
}

// RemoveRefreshToken deletes a refresh token from the user's list
func (u *User) RemoveRefreshToken(token string) error {
	for i, t := range u.Account.RefreshTokens {
		if t == token {
			u.Account.RefreshTokens = append(u.Account.RefreshTokens[:i], u.Account.RefreshTokens[i+1:]...)
			return nil
		}
	}
	return errors.New("token was missing")
}

// Timestamps describes metadata for the User
type Timestamps struct {
	LastTokenRefresh int64 `bson:"lastTokenRefresh"`
	LastLogin        int64 `bson:"lastLogin"`
	CreatedAt        int64 `bson:"createdAt"`
	UpdatedAt        int64 `bson:"updatedAt"`
}

// ToAPI converts the object from DB to API format
func (o Timestamps) ToAPI() *api.User_Timestamps {
	return &api.User_Timestamps{
		LastTokenRefresh: o.LastTokenRefresh,
		LastLogin:        o.LastLogin,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}
