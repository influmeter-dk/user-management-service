package main

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// TODO: add http handler methods here, please avoid using direct DB access here, instead use data_methods.go to define wrapper functions

func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

func comparePasswordWithHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func checkEmailFormat(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	return re.MatchString(email)
}

func loginHandl(context *gin.Context) {

	// TODO: find user
	// TODO: compare password
	// TODO: check role
	// TODO: return success

	context.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func signupHandl(c *gin.Context) {

	user := c.MustGet("user").(User)

	if !checkEmailFormat(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email not valid"})
		return
	}

	password := hashPassword(user.Password)

	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password not valid"})
		return
	}

	user.Password = password

	id, err := CreateUser(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	c.JSON(http.StatusCreated, gin.H{"user_id": id})
}
