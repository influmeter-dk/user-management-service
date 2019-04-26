package main

import (
	"testing"
)

func TestPasswordHashingMethods(t *testing.T) {
	t.Run("try to hash empty string", func(t *testing.T) {
		hPw, err := hashPassword("")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		match, err := comparePasswordWithHash(hPw, "")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !match {
			t.Error("password should match hashed value")
		}
	})

	t.Run("hash and compare strings", func(t *testing.T) {
		hPw, err := hashPassword("testPassword")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		match, err := comparePasswordWithHash(hPw, "testPassword")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !match {
			t.Error("password should match hashed value")
		}
	})
}
