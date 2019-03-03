package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func instanceUserColRef(instanceID string) *mongo.Collection {
	return dbClient.Database(instanceID + "_users").Collection("users")
}

func createUserDB(instanceID string, user User) (id string, err error) {
	_, err = findUserByEmail(instanceID, user.Email)
	if err == nil {
		err = errors.New("user already exists")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	user.ObjectInfos.CreatedAt = time.Now().Unix()

	res, err := instanceUserColRef(instanceID).InsertOne(ctx, user)
	if err != nil {
		return
	}
	id = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func updateUserDB(instanceID string, updatedUser User) error {
	filter := bson.M{"_id": updatedUser.ID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	// Set last update time
	updatedUser.ObjectInfos.UpdatedAt = time.Now().Unix()

	newDoc := User{}
	rd := options.After
	fro := options.FindOneAndReplaceOptions{
		ReturnDocument: &rd,
	}
	err := instanceUserColRef(instanceID).FindOneAndReplace(ctx, filter, updatedUser, &fro).Decode(&newDoc)

	if err != nil {
		return err
	}
	if newDoc.ID != updatedUser.ID {
		return errors.New("no document found or updated")
	}
	return nil
}

func findUserByID(instanceID string, id string) (User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	elem := User{}
	err := instanceUserColRef(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func findUserByEmail(instanceID string, username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	elem := User{}
	filter := bson.M{"email": username}
	err := instanceUserColRef(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func updateUserPasswordDB(instanceID string, userID string, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"password": newPassword, "objectInfos.updatedAt": time.Now().Unix()}}
	_, err := instanceUserColRef(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func updateLoginTimeInDB(instanceID string, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"objectInfos.lastLogin": time.Now().Unix()}}
	_, err := instanceUserColRef(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func updateTokenRefreshTimeInDB(instanceID string, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"objectInfos.lastTokenRefresh": time.Now().Unix()}}
	_, err := instanceUserColRef(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func deleteUserDB(instanceID string, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()
	_, err := instanceUserColRef(instanceID).DeleteOne(ctx, filter, nil)
	return err
}
