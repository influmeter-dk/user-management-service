package main

import (
	"context"
	"errors"
	"regexp"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"

	influenzanet "github.com/Influenzanet/api/dist/go"
	user_api "github.com/Influenzanet/api/dist/go/user-management"
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

func checkPasswordFormat(password string) bool {
	return len(password) > 5
}

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*influenzanet.Status, error) {
	return nil, errors.New("not implemented")
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, creds *influenzanet.UserCredentials) (*user_api.UserAuthInfo, error) {
	if creds == nil {
		return nil, errors.New("invalid username and/or password")
	}
	user, err := FindUserByEmail(creds.Email)

	if err != nil {
		return nil, errors.New("invalid username and/or password")
	}

	if comparePasswordWithHash(user.Password, creds.Password) != nil {
		return nil, errors.New("invalid username and/or password")
	}

	if creds.LoginRole == "" {
		creds.LoginRole = "PARTICIPANT"
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

func (s *userManagementServer) SignupWithEmail(ctx context.Context, u *user_api.NewUser) (*user_api.UserAuthInfo, error) {
	if u == nil {
		return nil, errors.New("missing argument")
	}
	if !checkEmailFormat(u.Email) {
		return nil, errors.New("email not valid")
	}
	if !checkPasswordFormat(u.Password) {
		return nil, errors.New("password too weak")
	}

	password := hashPassword(u.Password)

	// Create user DB object from request:
	newUser := User{
		Email:    u.Email,
		Password: password,
		Roles:    []string{"PARTICIPANT"},
		// TODO: add profile
	}

	id, err := CreateUser(newUser)

	if err != nil {
		return nil, err
	}

	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &user_api.UserAuthInfo{
		UserId:            id,
		Roles:             newUser.Roles,
		AuthenticatedRole: newUser.Roles[0],
	}
	return response, nil
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *user_api.PasswordChangeMsg) (*influenzanet.Status, error) {
	if req == nil || req.Auth == nil {
		return nil, errors.New("missing argument")
	}

	if !checkPasswordFormat(req.NewPassword) {
		return nil, errors.New("new password too weak")
	}

	user, err := FindUserByID(req.Auth.UserId)
	if err != nil {
		return nil, errors.New("invalid user and/or password")
	}

	if comparePasswordWithHash(user.Password, req.OldPassword) != nil {
		return nil, errors.New("invalid user and/or password")
	}

	user.Password = hashPassword(req.NewPassword)

	err = UpdateUser(user)
	if err != nil {
		return nil, err
	}

	// TODO: initiate email notification for user about password update

	return &influenzanet.Status{
		Status: influenzanet.Status_NORMAL,
		Msg:    "password changed",
	}, nil
}
