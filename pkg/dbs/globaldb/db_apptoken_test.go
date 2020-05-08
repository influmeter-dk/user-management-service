package globaldb

import (
	"testing"

	"github.com/influenzanet/authentication-service/models"
)

func TestDbInterfaceMethodsForAppToken(t *testing.T) {
	appToken := models.AppToken{
		AppName:   "testapp",
		Instances: []string{testInstanceID},
		Tokens:    []string{"test1", "test2"},
	}
	ctx, cancel := getContext()
	defer cancel()

	_, err := collectionAppToken().InsertOne(ctx, appToken)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	t.Run("Find existing app token", func(t *testing.T) {
		res, err := findAppTokenInDB("test1")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if res.AppName != appToken.AppName {
			t.Error("app token object not retrieved correctly")
		}
	})

	t.Run("Try to find not existing app token", func(t *testing.T) {
		_, err := findAppTokenInDB("test3")
		if err == nil {
			t.Error("should not be found")
			return
		}
	})
}
