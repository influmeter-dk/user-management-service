package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/influenzanet/user-management-service/pkg/models"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port        string
	ServiceURLs struct {
		MessagingService string
		LoggingService   string
	}
	UserDBConfig                models.DBConfig
	GlobalDBConfig              models.DBConfig
	Intervals                   models.Intervals
	NewUserCountLimit           int64
	CleanUpUnverifiedUsersAfter int64
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv("USER_MANAGEMENT_LISTEN_PORT")
	conf.ServiceURLs.MessagingService = os.Getenv("ADDR_MESSAGING_SERVICE")
	conf.ServiceURLs.LoggingService = os.Getenv("ADDR_LOGGING_SERVICE")

	conf.UserDBConfig = getUserDBConfig()
	conf.GlobalDBConfig = getGlobalDBConfig()
	conf.Intervals = getIntervalsConfig()

	rl, err := strconv.Atoi(os.Getenv("NEW_USER_RATE_LIMIT"))
	if err != nil {
		log.Fatal("NEW_USER_RATE_LIMIT: " + err.Error())
	}
	conf.NewUserCountLimit = int64(rl)

	cleanUpThreshold, err := strconv.Atoi(os.Getenv("CLEAN_UP_UNVERIFIED_USERS_AFTER"))
	if err != nil {
		log.Fatal("CLEAN_UP_UNVERIFIED_USERS_AFTER: " + err.Error())
	}
	conf.CleanUpUnverifiedUsersAfter = int64(cleanUpThreshold)
	return conf
}

func getIntervalsConfig() models.Intervals {
	intervals := models.Intervals{
		TokenExpiryInterval:      time.Minute * time.Duration(defaultTokenExpirationMin),
		VerificationCodeLifetime: defaultVerificationCodeLifetime,
	}

	accessTokenExpiration, err := strconv.Atoi(os.Getenv(ENV_TOKEN_EXPIRATION_MIN))
	if err != nil {
		log.Println("using default token expiration")
	} else {
		intervals.TokenExpiryInterval = time.Minute * time.Duration(accessTokenExpiration)
	}

	newVerificationCodeLifetime, err := strconv.Atoi(os.Getenv(ENV_VERIFICATION_CODE_LIFETIME))
	if err != nil {
		log.Println("using default verification code lifetime")
	} else {
		intervals.VerificationCodeLifetime = int64(newVerificationCodeLifetime)
	}

	return intervals
}

func getUserDBConfig() models.DBConfig {
	connStr := os.Getenv("USER_DB_CONNECTION_STR")
	username := os.Getenv("USER_DB_USERNAME")
	password := os.Getenv("USER_DB_PASSWORD")
	prefix := os.Getenv("USER_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	noCursorTimeout := os.Getenv(ENV_USE_NO_CURSOR_TIMEOUT) == "true"

	DBNamePrefix := os.Getenv("DB_DB_NAME_PREFIX")

	return models.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		NoCursorTimeout: noCursorTimeout,
		MaxPoolSize:     MaxPoolSize,
		DBNamePrefix:    DBNamePrefix,
	}
}

func getGlobalDBConfig() models.DBConfig {
	connStr := os.Getenv("GLOBAL_DB_CONNECTION_STR")
	username := os.Getenv("GLOBAL_DB_USERNAME")
	password := os.Getenv("GLOBAL_DB_PASSWORD")
	prefix := os.Getenv("GLOBAL_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	DBNamePrefix := os.Getenv("DB_DB_NAME_PREFIX")

	return models.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		DBNamePrefix:    DBNamePrefix,
	}
}
