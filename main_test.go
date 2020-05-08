package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc/status"
)

var testInstanceID = strconv.FormatInt(time.Now().Unix(), 10)

func dropTestDB() {

	/*log.Println("Drop test database")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbClient.Database(conf.DB.DBNamePrefix + testInstanceID + "_users").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}*/
}

// Pre-Test Setup
func TestMain(m *testing.M) {
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}

func shouldHaveGrpcErrorStatus(err error, expectedError string) (bool, string) {
	if err == nil {
		return false, "should return an error"
	}
	st, ok := status.FromError(err)
	if !ok || st == nil {
		return false, fmt.Sprintf("unexpected error: %s", err.Error())
	}

	if len(expectedError) > 0 && st.Message() != expectedError {
		return false, fmt.Sprintf("wrong error: %s", st.Message())
	}
	return true, ""
}
