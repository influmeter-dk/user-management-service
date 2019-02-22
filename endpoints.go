package main

import (
	"context"
	"regexp"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, creds *influenzanet.UserCredentials) (*user_api.UserAuthInfo, error) {
	if creds == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	instanceID := creds.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}
	user, err := findUserByEmail(instanceID, creds.Email)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if comparePasswordWithHash(user.Password, creds.Password) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	response := &user_api.UserAuthInfo{
		UserId:     user.ID.Hex(),
		Roles:      user.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, u *user_api.NewUser) (*user_api.UserAuthInfo, error) {
	if u == nil || u.Auth == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !checkEmailFormat(u.Auth.Email) {
		return nil, status.Error(codes.InvalidArgument, "email not valid")
	}
	if !checkPasswordFormat(u.Auth.Password) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password := hashPassword(u.Auth.Password)

	// Create user DB object from request:
	newUser := User{
		Email:    u.Auth.Email,
		Password: password,
		Roles:    []string{"PARTICIPANT"},
		// TODO: add profile
	}

	instanceID := u.Auth.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}

	id, err := createUserDB(instanceID, newUser)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &user_api.UserAuthInfo{
		UserId:     id,
		Roles:      newUser.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *user_api.PasswordChangeMsg) (*influenzanet.Status, error) {
	if req == nil || req.Auth == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !checkPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "new password too weak")
	}

	user, err := findUserByID(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	if comparePasswordWithHash(user.Password, req.OldPassword) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	user.Password = hashPassword(req.NewPassword)

	err = updateUserDB(req.Auth.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: initiate email notification for user about password update

	return &influenzanet.Status{
		Status: influenzanet.Status_NORMAL,
		Msg:    "password changed",
	}, nil
}
