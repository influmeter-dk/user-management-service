package globaldb

import (
	"github.com/influenzanet/user-management-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) FindAppToken(token string) (appTokenInfos models.AppToken, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"tokens": token}
	err = dbService.collectionAppToken().FindOne(ctx, filter).Decode(&appTokenInfos)
	return
}
