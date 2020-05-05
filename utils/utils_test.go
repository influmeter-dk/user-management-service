package utils

import (
	"log"
	"testing"

	"github.com/influenzanet/user-management-service/api"
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
