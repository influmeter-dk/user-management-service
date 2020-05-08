package globaldb

import (
	"errors"

	"github.com/influenzanet/authentication-service/models"
	"github.com/influenzanet/authentication-service/tokens"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) AddTempToken(t models.TempToken) (token string, err error) {
	ctx, cancel := getContext()
	defer cancel()

	t.Token, err = tokens.GenerateUniqueTokenString()
	if err != nil {
		return token, err
	}

	_, err = collectionRefTempToken().InsertOne(ctx, t)
	if err != nil {
		return token, err
	}
	token = t.Token
	return
}

func (dbService *GlobalDBService) GetTempTokenForUser(instanceID string, uid string, purpose string) (tokens models.TempTokens, err error) {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"instanceID": instanceID, "userID": uid}
	if len(purpose) > 0 {
		filter["purpose"] = purpose
	}

	cur, err := collectionRefTempToken().Find(ctx, filter)
	if err != nil {
		return tokens, err
	}
	defer cur.Close(ctx)

	tokens = []models.TempToken{}
	for cur.Next(ctx) {
		var result models.TempToken
		err := cur.Decode(&result)
		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, result)
	}
	if err := cur.Err(); err != nil {
		return tokens, err
	}
	return tokens, nil
}

func (dbService *GlobalDBService) GetTempToken(token string) (models.TempToken, error) {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"token": token}

	t := models.TempToken{}
	err := collectionRefTempToken().FindOne(ctx, filter).Decode(&t)
	return t, err
}

func (dbService *GlobalDBService) DeleteTempToken(token string) error {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"token": token}
	res, err := collectionRefTempToken().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("document not found")
	}
	return nil
}

func (dbService *GlobalDBService) DeleteAllTempTokenForUser(instanceID string, userID string, purpose string) error {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"instanceID": instanceID, "userID": userID}
	if len(purpose) > 0 {
		filter["purpose"] = purpose
	}
	res, err := collectionRefTempToken().DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("document not found")
	}
	return nil
}
