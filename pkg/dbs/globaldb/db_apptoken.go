package globaldb

import (
	"github.com/influenzanet/authentication-service/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) FindAppToken(token string) (appTokenInfos models.AppToken, err error) {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"tokens": token}
	err = collectionAppToken().FindOne(ctx, filter).Decode(&appTokenInfos)
	return
}
