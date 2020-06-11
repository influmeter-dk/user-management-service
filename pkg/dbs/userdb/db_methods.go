package userdb

import (
	"errors"
	"log"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *UserDBService) AddUser(instanceID string, user models.User) (id string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"account.accountID": user.Account.AccountID}
	upsert := true
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}
	res, err := dbService.collectionRefUsers(instanceID).UpdateOne(ctx, filter, bson.M{
		"$setOnInsert": user,
	}, &opts)
	if err != nil {
		return
	}

	if res.UpsertedCount < 1 {
		err = errors.New("user already exists")
		return
	}

	id = res.UpsertedID.(primitive.ObjectID).Hex()
	return
}

// low level find and replace
func (dbService *UserDBService) _updateUserInDB(orgID string, user models.User) (models.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	elem := models.User{}
	filter := bson.M{"_id": user.ID}
	rd := options.After
	fro := options.FindOneAndReplaceOptions{
		ReturnDocument: &rd,
	}
	err := dbService.collectionRefUsers(orgID).FindOneAndReplace(ctx, filter, user, &fro).Decode(&elem)
	return elem, err
}

func (dbService *UserDBService) UpdateUser(instanceID string, updatedUser models.User) (models.User, error) {
	// Set last update time
	updatedUser.Timestamps.UpdatedAt = time.Now().Unix()
	return dbService._updateUserInDB(instanceID, updatedUser)
}

func (dbService *UserDBService) GetUserByID(instanceID string, id string) (models.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := dbService.getContext()
	defer cancel()

	elem := models.User{}
	err := dbService.collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func (dbService *UserDBService) GetUserByAccountID(instanceID string, username string) (models.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	elem := models.User{}
	filter := bson.M{"account.accountID": username}
	err := dbService.collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func (dbService *UserDBService) UpdateUserPassword(instanceID string, userID string, newPassword string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"account.password": newPassword, "timestamps.updatedAt": time.Now().Unix()}}
	_, err := dbService.collectionRefUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (dbService *UserDBService) UpdateAccountPreferredLang(instanceID string, userID string, lang string) (models.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}

	elem := models.User{}

	rd := options.After
	fro := options.FindOneAndUpdateOptions{
		ReturnDocument: &rd,
	}
	update := bson.M{"$set": bson.M{"account.preferredLanguage": lang, "timestamps.updatedAt": time.Now().Unix()}}
	err := dbService.collectionRefUsers(instanceID).FindOneAndUpdate(ctx, filter, update, &fro).Decode(&elem)
	return elem, err
}

func (dbService *UserDBService) UpdateContactPreferences(instanceID string, userID string, prefs models.ContactPreferences) (models.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}

	elem := models.User{}

	rd := options.After
	fro := options.FindOneAndUpdateOptions{
		ReturnDocument: &rd,
	}
	update := bson.M{"$set": bson.M{"contactPreferences": prefs, "timestamps.updatedAt": time.Now().Unix()}}
	err := dbService.collectionRefUsers(instanceID).FindOneAndUpdate(ctx, filter, update, &fro).Decode(&elem)
	return elem, err
}

func (dbService *UserDBService) UpdateLoginTime(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"timestamps.lastLogin": time.Now().Unix()}}
	_, err := dbService.collectionRefUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (dbService *UserDBService) DeleteUser(instanceID string, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := dbService.getContext()
	defer cancel()
	res, err := dbService.collectionRefUsers(instanceID).DeleteOne(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no user found with the given id")
	}
	return nil
}

func (dbService *UserDBService) FindNonParticipantUsers(instanceID string) (users []models.User, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"roles": bson.M{"$elemMatch": bson.M{"$in": bson.A{"RESEARCHER", "ADMIN"}}},
	}
	cur, err := dbService.collectionRefUsers(instanceID).Find(
		ctx,
		filter,
	)

	if err != nil {
		return users, err
	}
	defer cur.Close(ctx)

	users = []models.User{}
	for cur.Next(ctx) {
		var result models.User
		err := cur.Decode(&result)
		if err != nil {
			return users, err
		}

		users = append(users, result)
	}
	if err := cur.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func (dbService *UserDBService) PerfomActionForUsers(
	instanceID string,
	cbk func(instanceID string, user models.User, args ...interface{}) error,
	args ...interface{},
) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	cur, err := dbService.collectionRefUsers(instanceID).Find(
		ctx,
		filter,
	)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result models.User
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		if err := cbk(instanceID, result, args...); err != nil {
			log.Printf("PerfomActionForUsers: %v", err)
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}
