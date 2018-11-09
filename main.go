package main

import (
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

	log.Println(dbCreds)
	log.Println(conf.DbAddress)

	/*
		var err error
		dbClient, err = mongo.NewClient("mongodb://testttt:bar@localhost:27017")
		if err != nil {
			log.Fatal(err)
		}

		context.Background()
	*/
}

func init() {
	readConfig()
	dbInit()
	gin.SetMode(gin.ReleaseMode)

}

func main() {
	log.Println(conf.DbCredentialsPath)
	log.Println("Hello World")
	/*
		router := gin.Default()

		v1 := router.Group("/v1")
		{
			v1.POST("/login", nil)
			v1.POST("/signup", nil)
		}

		log.Fatal(router.Run(":3100"))
	*/
}
