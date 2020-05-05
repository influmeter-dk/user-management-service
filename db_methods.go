package main

import (
	"errors"
	"time"

	"github.com/influenzanet/user-management-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addUserToDB(instanceID string, user models.User) (id string, err error) {
	if user.Account.Type == "email" {
		_, err = getUserByEmailFromDB(instanceID, user.Account.AccountID)
		if err == nil {
			err = errors.New("user already exists")
			return
		}
	}

	ctx, cancel := getContext()
	defer cancel()

	user.Timestamps.CreatedAt = time.Now().Unix()

	res, err := collectionRefUsers(instanceID).InsertOne(ctx, user)
	if err != nil {
		return
	}
	id = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

// low level find and replace
func _updateUserInDB(orgID string, user models.User) (models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	elem := models.User{}
	filter := bson.M{"_id": user.ID}
	rd := options.After
	fro := options.FindOneAndReplaceOptions{
		ReturnDocument: &rd,
	}
	err := collectionRefUsers(orgID).FindOneAndReplace(ctx, filter, user, &fro).Decode(&elem)
	return elem, err
}

func updateUserInDB(instanceID string, updatedUser models.User) (models.User, error) {
	// Set last update time
	updatedUser.Timestamps.UpdatedAt = time.Now().Unix()
	return _updateUserInDB(instanceID, updatedUser)
}

func getUserByIDFromDB(instanceID string, id string) (models.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := getContext()
	defer cancel()

	elem := models.User{}
	err := collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func getUserByEmailFromDB(instanceID string, username string) (models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	elem := models.User{}
	filter := bson.M{"account.accountID": username}
	err := collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func updateUserPasswordInDB(instanceID string, userID string, newPassword string) error {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"account.password": newPassword, "timestamps.updatedAt": time.Now().Unix()}}
	_, err := collectionRefUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func updateAccountPreferredLangDB(instanceID string, userID string, lang string) (models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}

	elem := models.User{}

	rd := options.After
	fro := options.FindOneAndUpdateOptions{
		ReturnDocument: &rd,
	}
	update := bson.M{"$set": bson.M{"account.preferredLanguage": lang, "timestamps.updatedAt": time.Now().Unix()}}
	err := collectionRefUsers(instanceID).FindOneAndUpdate(ctx, filter, update, &fro).Decode(&elem)
	return elem, err
}

func updateContactPreferencesDB(instanceID string, userID string, prefs models.ContactPreferences) (models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}

	elem := models.User{}

	rd := options.After
	fro := options.FindOneAndUpdateOptions{
		ReturnDocument: &rd,
	}
	update := bson.M{"$set": bson.M{"contactPreferences": prefs, "timestamps.updatedAt": time.Now().Unix()}}
	err := collectionRefUsers(instanceID).FindOneAndUpdate(ctx, filter, update, &fro).Decode(&elem)
	return elem, err
}

func updateLoginTimeInDB(instanceID string, id string) error {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"timestamps.lastLogin": time.Now().Unix()}}
	_, err := collectionRefUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func deleteUserFromDB(instanceID string, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := getContext()
	defer cancel()
	res, err := collectionRefUsers(instanceID).DeleteOne(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no user found with the given id")
	}
	return nil
}
