package models

import (
	"github.com/influenzanet/user-management-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TempToken is a database entry for a temporary token
type TempToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"token_id,omitempty"`
	Token      string             `bson:"token" json:"token"`
	Expiration int64              `bson:"expiration" json:"expiration"`
	Purpose    string             `bson:"purpose" json:"purpose"`
	UserID     string             `bson:"userID" json:"userID"`
	Info       string             `bson:"info" json:"info"`
	InstanceID string             `bson:"instanceID" json:"instanceID"`
}

// ToAPI converts the object from DB to API format
func (t TempToken) ToAPI() *api.TempTokenInfo {
	return &api.TempTokenInfo{
		Token:      t.Token,
		Expiration: t.Expiration,
		Purpose:    t.Purpose,
		UserId:     t.UserID,
		Info:       t.Info,
		InstanceId: t.InstanceID,
	}
}

// TempTokens is an array of TempToken
type TempTokens []TempToken

// ToAPI converts from DB formate into API format
func (items TempTokens) ToAPI() *api.TempTokenInfos {
	res := make([]*api.TempTokenInfo, len(items))
	for i, v := range items {
		res[i] = v.ToAPI()
	}

	return &api.TempTokenInfos{
		TempTokens: res,
	}
}
