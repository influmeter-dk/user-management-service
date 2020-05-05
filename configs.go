package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port string
	DB   struct {
		URI             string
		DBNamePrefix    string
		Timeout         int
		MaxPoolSize     uint64
		IdleConnTimeout int
	}
	ServiceURLs struct {
		AuthService string
	}
}

func initConfig() {
	conf.Port = os.Getenv("USER_MANAGEMENT_LISTEN_PORT")
	conf.ServiceURLs.AuthService = os.Getenv("ADDR_AUTH_SERVICE")
	getDBConfig()
}

func getDBConfig() {
	connStr := os.Getenv("USER_DB_CONNECTION_STR")
	username := os.Getenv("USER_DB_USERNAME")
	password := os.Getenv("USER_DB_PASSWORD")
	prefix := os.Getenv("USER_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	conf.DB.URI = fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	conf.DB.Timeout, err = strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	conf.DB.IdleConnTimeout, err = strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	conf.DB.MaxPoolSize = uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	conf.DB.DBNamePrefix = os.Getenv("DB_DB_NAME_PREFIX")
}
