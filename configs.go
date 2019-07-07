package main

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Port        int
	DB          dbConf
	ServiceURLs struct {
		AuthService string
	}
}

type dbConf struct {
	CredentialsPath string `yaml:"credentials_path"`
	Address         string `yaml:"address"`
	Timeout         int    `yaml:"timeout"`
}

type dbCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func readConfig() {
	data, err := ioutil.ReadFile("./configs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		log.Fatal(err)
	}
	conf.ServiceURLs.AuthService = os.Getenv("AUTH_SERVICE")
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
