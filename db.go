package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collections
func collectionRefUsers(instanceID string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_users").Collection("users")
}

// DB utils
func getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
}

// Connect to DB
func dbInit() {
	var err error
	dbClient, err = mongo.NewClient(
		options.Client().ApplyURI(conf.DB.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(conf.DB.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(conf.DB.MaxPoolSize),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := getContext()
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctx, conCancel := getContext()
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		log.Fatal("fail to connect to DB: " + err.Error())
	}
}
