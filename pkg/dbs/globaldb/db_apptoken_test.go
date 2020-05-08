package globaldb

import (
	"testing"

	"github.com/influenzanet/user-management-service/pkg/models"
)

func TestDbInterfaceMethodsForAppToken(t *testing.T) {
	appToken := models.AppToken{
		AppName:   "testapp",
		Instances: []string{testInstanceID},
		Tokens:    []string{"test1", "test2"},
	}
	ctx, cancel := testDBService.getContext()
	defer cancel()

	_, err := testDBService.collectionAppToken().InsertOne(ctx, appToken)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	t.Run("Find existing app token", func(t *testing.T) {
		res, err := testDBService.FindAppToken("test1")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if res.AppName != appToken.AppName {
			t.Error("app token object not retrieved correctly")
		}
	})

	t.Run("Try to find not existing app token", func(t *testing.T) {
		_, err := testDBService.FindAppToken("test3")
		if err == nil {
			t.Error("should not be found")
			return
		}
	})
}
