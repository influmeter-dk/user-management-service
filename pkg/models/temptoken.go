package models

import (
	"github.com/influenzanet/go-utils/pkg/api_types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TempToken is a database entry for a temporary token
type TempToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"token_id,omitempty"`
	Token      string             `bson:"token" json:"token"`
	Expiration int64              `bson:"expiration" json:"expiration"`
	Purpose    string             `bson:"purpose" json:"purpose"`
	UserID     string             `bson:"userID" json:"userID"`
	Info       map[string]string  `bson:"info" json:"info"`
	InstanceID string             `bson:"instanceID" json:"instanceID"`
}

// ToAPI converts the object from DB to API format
func (t *TempToken) ToAPI() *api_types.TempTokenInfo {
	if t == nil {
		return nil
	}
	return &api_types.TempTokenInfo{
		Token:      t.Token,
		Expiration: t.Expiration,
		Purpose:    t.Purpose,
		UserId:     t.UserID,
		Info:       t.Info,
		InstanceId: t.InstanceID,
	}
}

func TempTokenFromAPI(t *api_types.TempTokenInfo) *TempToken {
	if t == nil {
		return nil
	}
	return &TempToken{
		Token:      t.Token,
		Expiration: t.Expiration,
		Purpose:    t.Purpose,
		UserID:     t.UserId,
		Info:       t.Info,
		InstanceID: t.InstanceId,
	}
}

// TempTokens is an array of TempToken
type TempTokens []TempToken

// ToAPI converts from DB formate into API format
func (items TempTokens) ToAPI() *api_types.TempTokenInfos {
	res := make([]*api_types.TempTokenInfo, len(items))
	for i, v := range items {
		res[i] = v.ToAPI()
	}

	return &api_types.TempTokenInfos{
		TempTokens: res,
	}
}
