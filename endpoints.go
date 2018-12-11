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

func loginHandl(c *gin.Context) {

	u := c.MustGet("user").(User)

	user, err := FindUserByEmail(u.Email)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username and/or password"})
		return
	}

	if comparePasswordWithHash(user.Password, u.Password) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username and/or password"})
		return
	}

	if !user.HasRole(u.Role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "missing required role"})
		return
	}

	response := &UserLoginResponse{
		ID:   user.ID.Hex(),
		Role: u.Role,
	}

	c.JSON(http.StatusOK, response)
}

func signupHandl(c *gin.Context) {

	u := c.MustGet("user").(User)

	if !checkEmailFormat(u.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email not valid"})
		return
	}

	password := hashPassword(u.Password)

	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password not valid"})
		return
	}

	u.Password = password
	u.Roles = []string{"PARTICIPANT"}

	id, err := CreateUser(u)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &UserLoginResponse{
		ID:   id,
		Role: "PARTICIPANT",
	}

	c.JSON(http.StatusCreated, response)
}
