package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func instanceDBRef(instanceID string) *mongo.Collection {
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

	res, err := instanceDBRef(instanceID).InsertOne(ctx, user)
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

	newDoc := User{}
	err := instanceDBRef(instanceID).FindOneAndReplace(ctx, filter, updatedUser, nil).Decode(&newDoc)

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
	err := instanceDBRef(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func findUserByEmail(instanceID string, username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	elem := User{}
	filter := bson.M{"email": username}
	err := instanceDBRef(instanceID).FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func deleteUserDB(instanceID string, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()
	_, err := instanceDBRef(instanceID).DeleteOne(ctx, filter, nil)
	return err
}
