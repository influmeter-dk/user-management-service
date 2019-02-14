package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/mongo"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"

	user_api "github.com/Influenzanet/api/user-management"
)

type config struct {
	Port              int    `yaml:"port"`
	DbCredentialsPath string `yaml:"db_credentials_path"`
	DbAddress         string `yaml:"db_address"`
	DbTimeout         int    `yaml:"db_timeout"`
}

type dbCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type userManagementServer struct {
}

var dbClient *mongo.Client
var userCollection *mongo.Collection
var conf config

func readConfig() {
	data, err := ioutil.ReadFile("./configs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		log.Fatal(err)
	}
}

func readDBcredentials(path string) (dbCredentials, error) {
	var dbCreds dbCredentials
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return dbCreds, err
	}
	err = yaml.Unmarshal([]byte(data), &dbCreds)
	if err != nil {
		return dbCreds, err
	}
	return dbCreds, nil
}

func dbInit() {
	dbCreds, err := readDBcredentials(conf.DbCredentialsPath)
	if err != nil {
		log.Fatal(err)
	}

	// mongodb+srv://user-management-service:<PASSWORD>@influenzanettestdbcluster-pwvbz.mongodb.net/test?retryWrites=true
	address := fmt.Sprintf(`mongodb+srv://%s:%s@%s`, dbCreds.Username, dbCreds.Password, conf.DbAddress)

	dbClient, err = mongo.NewClient(address)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DbTimeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	userCollection = dbClient.Database("users").Collection("users")
}

func init() {
	readConfig()
	dbInit()
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	user_api.RegisterUserManagementApiServer(grpcServer, &userManagementServer{})
	grpcServer.Serve(lis)
}
