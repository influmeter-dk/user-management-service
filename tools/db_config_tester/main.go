package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/go-utils/pkg/constants"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const INSTANCE_ID = "v1"

var userDB *userdb.UserDBService

func init() {
	conf := getDBConfig()
	userDB = userdb.NewUserDBService(conf)
}

func main() {
	parseFlags()
}

func getDBConfig() models.DBConfig {
	connStr := os.Getenv("USER_DB_CONNECTION_STR")
	username := os.Getenv("USER_DB_USERNAME")
	password := os.Getenv("USER_DB_PASSWORD")
	prefix := os.Getenv("USER_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		logger.Error.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		logger.Error.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	noCursorTimeout := os.Getenv("ENV_USE_NO_CURSOR_TIMEOUT") == "true"

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

func generateRandomUser() models.User {
	dummyID, _ := tokens.GenerateUniqueTokenString()
	return models.User{
		Account: models.Account{
			Type:               "email",
			AccountID:          dummyID,
			AccountConfirmedAt: time.Now().Unix(), // not confirmed yet
			Password:           dummyID,
			PreferredLanguage:  "no",
		},
		Roles: []string{constants.USER_ROLE_ADMIN},
		Profiles: []models.Profile{
			{
				ID:                 primitive.NewObjectID(),
				AvatarID:           "default",
				ConsentConfirmedAt: time.Now().Unix(),
				MainProfile:        true,
			},
		},
		Timestamps: models.Timestamps{
			CreatedAt: time.Now().Unix(),
		},
	}
}

func generateUsers(usersToGenerate int) {
	for a := 0; a < usersToGenerate; a++ {
		_, err := userDB.AddUser(INSTANCE_ID, generateRandomUser())
		if err != nil {
			logger.Error.Fatal(err)
		}
	}
	logger.Info.Printf("%v users generated", usersToGenerate)
}

func parseFlags() {
	usersToGenerate := flag.Int("generateUsers", -1, "How many users to generate.")
	flag.Parse()

	if *usersToGenerate > 0 {
		generateUsers(*usersToGenerate)
	}
}
