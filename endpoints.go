package main

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"

	influenzanet "github.com/Influenzanet/api"
	user_api "github.com/Influenzanet/api/user-management"
)

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

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*influenzanet.Status, error) {
	return nil, errors.New("not implemented")
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, creds *influenzanet.UserCredentials) (*user_api.UserAuthInfo, error) {
	user, err := FindUserByEmail(creds.Email)

	if err != nil {
		return nil, errors.New("invalid username and/or password")
	}

	if comparePasswordWithHash(user.Password, creds.Password) != nil {
		return nil, errors.New("invalid username and/or password")
	}

	if !user.HasRole(creds.LoginRole) {
		return nil, errors.New("missing required role")
	}

	response := &user_api.UserAuthInfo{
		UserId:            user.ID.Hex(),
		Roles:             user.Roles,
		AuthenticatedRole: creds.LoginRole,
	}
	return response, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, creds *user_api.NewUser) (*user_api.UserAuthInfo, error) {
	return nil, errors.New("not implemented")
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *user_api.PasswordChangeMsg) (*influenzanet.Status, error) {
	return nil, errors.New("not implemented")
}

func signupHandl(c *gin.Context) {

	u := c.MustGet("user").(User)

	if !checkEmailFormat(u.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email not valid"})
		return
	}

	password := hashPassword(u.Password)

	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing password"})
		return
	}

	u.Password = password
	u.Roles = []string{"PARTICIPANT"}

	id, err := CreateUser(u)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &UserLoginResponse{
		ID:                id,
		Roles:             u.Roles,
		AuthenticatedRole: u.Role,
	}

	c.JSON(http.StatusCreated, response)
}

func passwordChangeHandl(c *gin.Context) {
	u := c.MustGet("user").(User)

	user, err := FindUserByEmail(u.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid username and/or password"})
		return
	}

	if comparePasswordWithHash(user.Password, u.Password) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username and/or password"})
		return
	}

	if u.NewPassword == "" || u.NewPasswordRepeat == "" || u.NewPassword != u.NewPasswordRepeat {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}

	password := hashPassword(u.NewPassword)

	if password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing new password"})
		return
	}

	user.Password = password

	err = UpdateUser(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
