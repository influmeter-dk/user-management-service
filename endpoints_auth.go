package main

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/influenzanet/user-management-service/api"
	utils "github.com/influenzanet/user-management-service/utils"
)

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*api.Status, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, creds *api.UserCredentials) (*api.UserAuthInfo, error) {
	if creds == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	instanceID := creds.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}
	user, err := getUserByEmailFromDB(instanceID, creds.Email)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	match, err := utils.ComparePasswordWithHash(user.Account.Password, creds.Password)
	if err != nil || !match {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if err := updateLoginTimeInDB(instanceID, user.ID.Hex()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response := &api.UserAuthInfo{
		UserId:     user.ID.Hex(),
		Roles:      user.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, u *api.UserCredentials) (*api.UserAuthInfo, error) {
	if u == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !utils.CheckEmailFormat(u.Email) {
		return nil, status.Error(codes.InvalidArgument, "email not valid")
	}
	if !utils.CheckPasswordFormat(u.Password) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password, err := utils.HashPassword(u.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create user DB object from request:
	newUser := User{
		Account: Account{
			Type:           "email",
			Email:          u.Email,
			EmailConfirmed: false,
			Password:       password,
		},
		Roles: []string{"PARTICIPANT"},
	}

	instanceID := u.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}

	id, err := addUserToDB(instanceID, newUser)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Println("new user created")
	// TODO: generate email confirmation token
	// TODO: send email with confirmation request

	response := &api.UserAuthInfo{
		UserId:     id,
		Roles:      newUser.Roles,
		InstanceId: instanceID,
	}
	return response, nil
}

func (s *userManagementServer) CheckRefreshToken(ctx context.Context, req *api.UserReference) (*api.Status, error) {
	if req == nil || req.Token == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) TokenRefreshed(ctx context.Context, req *api.UserReference) (*api.Status, error) {
	if req == nil || req.Token == "" || req.UserId == "" || req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}
	user.AddRefreshToken(req.Token)
	user.ObjectInfos.LastTokenRefresh = time.Now().Unix()

	user, err = updateUserInDB(req.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "token refresh time updated",
	}, nil
}
