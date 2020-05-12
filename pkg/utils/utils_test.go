package utils

import (
	"log"
	"strings"
	"testing"

	"github.com/influenzanet/user-management-service/pkg/api"
)

func TestCheckPasswordFormat(t *testing.T) {
	log.Println("RUN util test")
	t.Run("with a too short password", func(t *testing.T) {
		if CheckPasswordFormat("1n34T6@") {
			t.Error("should be false")
		}
	})
	t.Run("with a too weak password", func(t *testing.T) {
		if CheckPasswordFormat("13342678") {
			t.Error("should be false")
		}
	})
	t.Run("with good passwords", func(t *testing.T) {
		if !CheckPasswordFormat("11111aaaa") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("1n34T678") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("nnnnnnT@@") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("TTTTTTTT77.") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("Tt1,.Lo%4") {
			t.Error("should be true")
		}
	})
}

func TestCheckEmailFormat(t *testing.T) {
	t.Run("with missing @", func(t *testing.T) {
		if CheckEmailFormat("t.t.t") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong domain format", func(t *testing.T) {
		if CheckEmailFormat("t@t") {
			t.Error("should be false")
		}
		if CheckEmailFormat("t@t.") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong local format", func(t *testing.T) {
		if CheckEmailFormat("@t.t") {
			t.Error("should be false")
		}
	})

	t.Run("with too many @", func(t *testing.T) {
		if CheckEmailFormat("t@@t.t") {
			t.Error("should be false")
		}
	})
	t.Run("with correct format", func(t *testing.T) {
		if !CheckEmailFormat("t@t.t") {
			t.Error("should be true")
		}
	})
}

func TestIsTokenEmpty(t *testing.T) {
	t.Run("check with nil input", func(t *testing.T) {
		if !IsTokenEmpty(nil) {
			t.Error("should be true")
		}
	})

	t.Run("check with empty id", func(t *testing.T) {
		if !IsTokenEmpty(&api.TokenInfos{Id: "", InstanceId: "testid"}) {
			t.Error("should be true")
		}
	})

	t.Run("check with empty InstanceId", func(t *testing.T) {
		if !IsTokenEmpty(&api.TokenInfos{InstanceId: "", Id: "testid"}) {
			t.Error("should be true")
		}
	})

	t.Run("check with not empty id", func(t *testing.T) {
		if IsTokenEmpty(&api.TokenInfos{Id: "testid", InstanceId: "testid"}) {
			t.Error("should be false")
		}
	})
}

func TestCheckRoleInToken(t *testing.T) {
	t.Run("check with nil input", func(t *testing.T) {
		if CheckRoleInToken(nil, "") {
			t.Error("should be false")
		}
	})

	t.Run("check with no payload ", func(t *testing.T) {
		tokenInf := &api.TokenInfos{
			Id: "testid",
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with single role - wrong", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"notthesame"}, ",")

		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with single role - right", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"testrole"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if !CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be true")
		}
	})

	t.Run("check with multiple roles - wrong", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"r1", "r2", "r4"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be false")
		}
	})

	t.Run("check with multiple roles - right", func(t *testing.T) {
		payload := map[string]string{}
		payload["roles"] = strings.Join([]string{"r1", "r2", "r4", "testrole"}, ",")
		tokenInf := &api.TokenInfos{
			Id:      "testid",
			Payload: payload,
		}
		if !CheckRoleInToken(tokenInf, "testrole") {
			t.Error("should be true")
		}
	})
}
