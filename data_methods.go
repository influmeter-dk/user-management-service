package main

import (
	"context"
	"errors"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

// TODO: add methods interfacing database here - this is an abstraction layer to the DB

func CreateUser(user User) (id string, err error) {
	_, err = FindUserByEmail(user.Email)
	if err == nil {
		err = errors.New("user already exists")
		return
	}

	res, err := userCollection.InsertOne(context.Background(), user)
	if err != nil {
		return
	}
	id = res.InsertedID.(objectid.ObjectID).Hex()
	return
}

func UpdateUser(updatedUser User) error {
	filter := bson.NewDocument(bson.EC.ObjectID("_id", updatedUser.ID))
	res := userCollection.FindOneAndReplace(context.Background(), filter, updatedUser, nil)
	newDoc := User{}
	err := res.Decode(&newDoc)
	if err != nil {
		return err
	}
	if newDoc.ID != updatedUser.ID {
		return errors.New("no document found or updated")
	}
	return nil
}

func FindUserByID(id string) (User, error) {
	_id, _ := objectid.FromHex(id)
	filter := bson.NewDocument(bson.EC.ObjectID("_id", _id))

	res := userCollection.FindOne(context.Background(), filter, nil)

	elem := User{}
	err := res.Decode(&elem)
	return elem, err
}

func FindUserByEmail(username string) (User, error) {
	filter := map[string]string{"email": username}
	res := userCollection.FindOne(context.Background(), filter, nil)

	elem := User{}
	err := res.Decode(&elem)

	return elem, err
}

func DeleteUser(id string) error {
	_id, _ := objectid.FromHex(id)
	filter := bson.NewDocument(bson.EC.ObjectID("_id", _id))
	_, err := userCollection.DeleteOne(context.Background(), filter, nil)
	return err
}
