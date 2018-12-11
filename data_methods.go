package main

import (
	"context"
	"errors"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// TODO: add methods interfacing database here - this is an abstraction layer to the DB
func CreateUser(user User) (id string, err error) {
	_, err = FindUserByEmail(user.Email)
	if err == nil {
		err = errors.New("user already exists")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()

	res, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		return
	}
	id = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func UpdateUser(updatedUser User) error {
	filter := bson.M{"_id": updatedUser.ID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()

	newDoc := User{}
	err := userCollection.FindOneAndReplace(ctx, filter, updatedUser, nil).Decode(&newDoc)

	if err != nil {
		return err
	}
	if newDoc.ID != updatedUser.ID {
		return errors.New("no document found or updated")
	}
	return nil
}

func FindUserByID(id string) (User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()

	elem := User{}
	err := userCollection.FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func FindUserByEmail(username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()

	elem := User{}
	filter := bson.M{"email": username}
	err := userCollection.FindOne(ctx, filter).Decode(&elem)

	return elem, err
}

func DeleteUser(id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()
	_, err := userCollection.DeleteOne(ctx, filter, nil)
	return err
}
