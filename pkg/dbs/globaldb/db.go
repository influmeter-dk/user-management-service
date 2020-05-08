package globaldb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GlobalDBService struct {
	DBClient     *mongo.Client
	timeout      int
	DBNamePrefix string
}

func NewGlobalDBService(
	URI string,
	timeout int,
	idleConnTimeout int,
	maxPoolSize uint64,
	DBNamePrefix string,
) *GlobalDBService {
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

	return &GlobalDBService{
		DBClient:     dbClient,
		timeout:      timeout,
		DBNamePrefix: DBNamePrefix,
	}
}

// Collections
func (dbService *GlobalDBService) collectionRefTempToken() *mongo.Collection {
	return dbClient.Database(dbService.DBNamePrefix + "global-infos").Collection("temp-tokens")
}

func (dbService *GlobalDBService) collectionAppToken() *mongo.Collection {
	return dbClient.Database(dbService.DBNamePrefix + "global-infos").Collection("app-tokens")
}
