package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/mongo"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	DbCredentialsPath string `yaml:"db_credentials_path"`
	DbAddress         string `yaml:"db_address"`
}

type dbCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

	err = dbClient.Connect(context.Background())
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
	/*
		log.Println("Hello World")

		currentUser := User{
			Email:    "test2@test.com",
			Password: HashPassword("testpassword"),
			Roles:    []string{"participant"},
		}
		id, err := CreateUser(currentUser)
		if err != nil {
			log.Println(err)
		}
		log.Println(id)

		user, _ := FindUserByEmail("test@test.com")

		log.Println(ComparePasswordWithHash(user.Password, "testpassword2"))
		log.Println(ComparePasswordWithHash(user.Password, "testpassword"))

		FindUserByID("5be84fb1c6dcde996d940385")

		nuser, err := FindUserByEmail("testuser2")
		if err != nil {
			log.Fatal(err)
		}
		log.Println(nuser)
	*/
	/*
		cur, err := collection.Find(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}
		defer cur.Close(context.Background())
		for cur.Next(context.Background()) {
			elem := bson.NewDocument()
			err := cur.Decode(elem)
			if err != nil {
				log.Fatal(err)
			}

			log.Println(elem)
			// do something with elem....
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
	*/

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/login", bindUserFromBodyMiddleware(), loginHandl)
		v1.POST("/signup", bindUserFromBodyMiddleware(), signupHandl)
	}

	log.Fatal(router.Run(":3200"))

}
