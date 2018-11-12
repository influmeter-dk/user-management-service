package main

import (
	"golang.org/x/crypto/bcrypt"
)

// TODO: add http handler methods here, please avoid using direct DB access here, instead use data_methods.go to define wrapper functions

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

func ComparePasswordWithHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
