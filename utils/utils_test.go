package utils

import (
	"log"
	"testing"
)

func TestCheckPasswordFormat(t *testing.T) {
	log.Println("RUN util test")
	t.Run("with a too short password", func(t *testing.T) {
		if CheckPasswordFormat("1n34T6@") {
			t.Error("should be false")
		}
	})
	t.Run("with a too weak password", func(t *testing.T) {
		if CheckPasswordFormat("1n34t678") {
			t.Error("should be false")
		}
	})
	t.Run("with good passwords", func(t *testing.T) {
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
