package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func loginHandl(context *gin.Context) {
	if context.Request.ContentLength == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "payload missing"})
		return
	}

	// TODO: parse body
	// TODO: find user
	// TODO: compare password
	// TODO: check role
	// TODO: return success
}

func signupHandl(context *gin.Context) {
	if context.Request.ContentLength == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "payload missing"})
		return
	}

	// TODO: parse body
	// TODO: hash password
	// TODO: create user
	// TODO: return success
}
