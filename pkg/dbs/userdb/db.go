package userdb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDBService struct {
	DBClient     *mongo.Client
	timeout      int
	DBNamePrefix string
}

func NewUserDBService(
	URI string,
	timeout int,
	idleConnTimeout int,
	maxPoolSize uint64,
	DBNamePrefix string,
) *UserDBService {
	var err error
	dbClient, err := mongo.NewClient(
		options.Client().ApplyURI(URI),
		options.Client().SetMaxConnIdleTime(time.Duration(idleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(maxPoolSize),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		log.Fatal("fail to connect to DB: " + err.Error())
	}

	return &UserDBService{
		DBClient:     dbClient,
		timeout:      timeout,
		DBNamePrefix: DBNamePrefix,
	}
}

// Collections
func (dbService *UserDBService) collectionRefUsers(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_users").Collection("users")
}

// DB utils
func (dbService *UserDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}
