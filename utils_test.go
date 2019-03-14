package main

import (
	"testing"
)

func TestCheckPasswordFormat(t *testing.T) {
	t.Run("with a too short password", func(t *testing.T) {
		if checkPasswordFormat("1n34T6@") {
			t.Error("should be false")
		}
	})
	t.Run("with a too weak password", func(t *testing.T) {
		if checkPasswordFormat("1n34t678") {
			t.Error("should be false")
		}
	})
	t.Run("with good passwords", func(t *testing.T) {
		if !checkPasswordFormat("1n34T678") {
			t.Error("should be true")
		}
		if !checkPasswordFormat("nnnnnnT@@") {
			t.Error("should be true")
		}
		if !checkPasswordFormat("TTTTTTTT77.") {
			t.Error("should be true")
		}
		if !checkPasswordFormat("Tt1,.Lo%4") {
			t.Error("should be true")
		}
	})
}

func TestCheckEmailFormat(t *testing.T) {
	t.Run("with missing @", func(t *testing.T) {
		if checkEmailFormat("t.t.t") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong domain format", func(t *testing.T) {
		if checkEmailFormat("t@t") {
			t.Error("should be false")
		}
		if checkEmailFormat("t@t.") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong local format", func(t *testing.T) {
		if checkEmailFormat("@t.t") {
			t.Error("should be false")
		}
	})

	t.Run("with too many @", func(t *testing.T) {
		if checkEmailFormat("t@@t.t") {
			t.Error("should be false")
		}
	})
	t.Run("with correct format", func(t *testing.T) {
		if !checkEmailFormat("t@t.t") {
			t.Error("should be true")
		}
	})
}

func TestHashAndComparePassword(t *testing.T) {
	testPw := "öl.§42342_,sdfsdf"
	testPwHash := hashPassword(testPw)

	t.Run("with wrong password", func(t *testing.T) {
		err := comparePasswordWithHash(testPwHash, testPw+"1")
		if err == nil {
			t.Error("wrong pasword should generate error")
		}
	})
	t.Run("with correct password", func(t *testing.T) {
		err := comparePasswordWithHash(testPwHash, testPw)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})
}
