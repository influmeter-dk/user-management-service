package globaldb

import (
	"log"
	"testing"
	"time"

	"github.com/influenzanet/authentication-service/models"
	"github.com/influenzanet/authentication-service/tokens"
)

// Testing Database Interface methods
func TestDbInterfaceMethodsForTempToken(t *testing.T) {
	testTempToken := models.TempToken{
		UserID:     "test_user_id",
		Purpose:    "test_purpose1",
		InstanceID: testInstanceID,
		Expiration: tokens.GetExpirationTime(10 * time.Second),
	}
	tokenStr := ""

	t.Run("Add temporary token to DB", func(t *testing.T) {
		ts, err := addTempTokenDB(testTempToken)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		tokenStr = ts

		testTempToken2 := testTempToken
		testTempToken2.Purpose = "test_purpose2"
		_, err = addTempTokenDB(testTempToken2)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	})

	t.Run("try to get temporary token by wrong token string", func(t *testing.T) {
		tempToken, err := getTempTokenFromDB(tokenStr + "++")
		if err == nil || tempToken.UserID != "" {
			t.Error(tempToken)
			t.Error("token should not be found")
			return
		}
	})

	t.Run("get temporary token by token string", func(t *testing.T) {
		tempToken, err := getTempTokenFromDB(tokenStr)
		if err != nil {
			t.Error("token not found by token string")
			return
		}
		if tempToken.UserID != testTempToken.UserID || tempToken.Purpose != testTempToken.Purpose || tempToken.Expiration != testTempToken.Expiration {
			t.Error("temp token does not match")
			t.Error(tempToken)
			return
		}
	})

	t.Run("try to get temporary token by wrong user id", func(t *testing.T) {
		tt, err := getTempTokenForUserDB(testInstanceID, testTempToken.UserID+"1", "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tt) > 0 {
			t.Error(tt)
			t.Error("token should not be found")
			return
		}
	})

	t.Run("try to get temporary token by wrong instace id", func(t *testing.T) {
		tt, err := getTempTokenForUserDB(testInstanceID+"1", testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tt) > 0 {
			t.Error(tt)
			t.Error("token should not be found")
			return
		}
	})

	t.Run("try to get temporary token by wrong purpose", func(t *testing.T) {
		tt, err := getTempTokenForUserDB(testInstanceID, testTempToken.UserID, testTempToken.Purpose+"1")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tt) > 0 {
			t.Error(tt)
			t.Error("token should not be found")
			return
		}
	})

	t.Run("get temporary token by user_id+instance_id", func(t *testing.T) {
		tt, err := getTempTokenForUserDB(testInstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		log.Println(tt)
		if len(tt) < 2 {
			t.Error("tokens should be found")
			return
		}
	})

	t.Run("get temporary token by user_id+instance_id+purpose", func(t *testing.T) {
		tt, err := getTempTokenForUserDB(testInstanceID, testTempToken.UserID, testTempToken.Purpose)
		if err != nil {
			t.Error(err)
			return
		}
		if len(tt) > 1 {
			t.Error("only one token should be found")
			return
		}
	})

	t.Run("Try delete not existing temporary token", func(t *testing.T) {
		err := deleteTempTokenDB(tokenStr + "1")
		if err == nil {
			t.Error("should not be deleted")
			return
		}
	})

	t.Run("Delete temporary token", func(t *testing.T) {
		err := deleteTempTokenDB(tokenStr)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = getTempTokenFromDB(testTempToken.Token)
		if err == nil {
			t.Error("token should be deleted by now")
			return
		}
		_, err = addTempTokenDB(testTempToken)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Delete all temporary token of a user_id with empty instance_id", func(t *testing.T) {
		err := deleteAllTempTokenForUserDB("", testTempToken.UserID, "")
		if err == nil {
			t.Error("should not be deleted, since instance id is missing")
			return
		}
	})

	t.Run("Try to delete all temporary token of a user_id with wrong id, correct instance_id", func(t *testing.T) {
		err := deleteAllTempTokenForUserDB(testTempToken.InstanceID, testTempToken.UserID+"3", "")
		if err == nil {
			t.Error("should not be deleted, because user id is wrong")
			return
		}
	})

	t.Run("Delete all temporary token of a user_id+instance_id+purpose", func(t *testing.T) {
		err := deleteAllTempTokenForUserDB(testTempToken.InstanceID, testTempToken.UserID, testTempToken.Purpose)
		if err != nil {
			t.Error(err)
			return
		}
		tokens, err := getTempTokenForUserDB(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tokens) != 1 {
			t.Error(tokens)
			t.Error("too many tokens found")
			return
		}
	})

	t.Run("Delete all temporary token of a user_id+instance_id", func(t *testing.T) {
		err := deleteAllTempTokenForUserDB(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		tokens, err := getTempTokenForUserDB(testTempToken.InstanceID, testTempToken.UserID, "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(tokens) > 0 {
			t.Error(tokens)
			t.Error("too many tokens found")
			return
		}
	})
}
