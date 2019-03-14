package main

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	influenzanet "github.com/influenzanet/api/dist/go"
	user_api "github.com/influenzanet/api/dist/go/user-management"
)



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

	if err := updateLoginTimeInDB(instanceID, user.ID.Hex()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response := &user_api.UserAuthInfo{
		UserId:     user.ID.Hex(),
		Roles:      user.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, u *influenzanet.UserCredentials) (*user_api.UserAuthInfo, error) {
	if u == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !checkEmailFormat(u.Email) {
		return nil, status.Error(codes.InvalidArgument, "email not valid")
	}
	if !checkPasswordFormat(u.Password) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password := hashPassword(u.Password)

	// Create user DB object from request:
	newUser := User{
		Email:    u.Email,
		Password: password,
		Roles:    []string{"PARTICIPANT"},
	}

	newUser.InitProfile()

	instanceID := u.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}

	id, err := createUserDB(instanceID, newUser)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Println("new user created")
	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &user_api.UserAuthInfo{
		UserId:     id,
		Roles:      newUser.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) TokenRefreshed(ctx context.Context, req *user_api.UserReference) (*influenzanet.Status, error) {
	if req == nil || req.Auth == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if err := updateTokenRefreshTimeInDB(req.Auth.InstanceId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &influenzanet.Status{
		Status: influenzanet.Status_NORMAL,
		Msg:    "token refresh time updated",
	}, nil
}

