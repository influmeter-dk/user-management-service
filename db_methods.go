package main

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addUserToDB(instanceID string, user User) (id string, err error) {
	if user.Account.Type == "email" {
		_, err = getUserByEmailFromDB(instanceID, user.Account.Email)
		if err == nil {
			err = errors.New("user already exists")
			return
		}
	}

	ctx, cancel := getContext()
	defer cancel()

	user.ObjectInfos.CreatedAt = time.Now().Unix()

	res, err := collectionRefUsers(instanceID).InsertOne(ctx, user)
	if err != nil {
		return
	}
	id = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

// low level find and replace
func _updateUserInDB(orgID string, user User) (User, error) {
	ctx, cancel := getContext()
	defer cancel()

	elem := User{}
	filter := bson.M{"_id": user.ID}
	rd := options.After
	fro := options.FindOneAndReplaceOptions{
		ReturnDocument: &rd,
	}
	err := collectionRefUsers(orgID).FindOneAndReplace(ctx, filter, user, &fro).Decode(&elem)
	return elem, err
}

func updateUserInDB(instanceID string, updatedUser User) (User, error) {
	// Set last update time
	updatedUser.ObjectInfos.UpdatedAt = time.Now().Unix()
	return _updateUserInDB(instanceID, updatedUser)
}

func getUserByIDFromDB(instanceID string, id string) (User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := getContext()
	defer cancel()

	elem := User{}
	err := collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func getUserByEmailFromDB(instanceID string, username string) (User, error) {
	ctx, cancel := getContext()
	defer cancel()

	elem := User{}
	filter := bson.M{"account.email": username}
	err := collectionRefUsers(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func updateUserPasswordInDB(instanceID string, userID string, newPassword string) error {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"account.password": newPassword, "objectInfos.updatedAt": time.Now().Unix()}}
	_, err := collectionRefUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func updateLoginTimeInDB(instanceID string, id string) error {
	ctx, cancel := getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"objectInfos.lastLogin": time.Now().Unix()}}
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
