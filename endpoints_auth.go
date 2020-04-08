package main

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	var username string
	if len(user.Roles) > 1 || len(user.Roles) == 1 && user.Roles[0] != "PARTICIPANT" {
		username = user.Account.Email
	}

	response := &api.UserAuthInfo{
		UserId:     user.ID.Hex(),
		Roles:      user.Roles,
		InstanceId: instanceID,
		Username:   username,
		ProfileId:  user.Profiles[0].ID.Hex(),
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
		Profiles: []Profile{
			Profile{
				ID: primitive.NewObjectID(),
			},
		},
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
	var username string
	if len(newUser.Roles) > 1 || len(newUser.Roles) == 1 && newUser.Roles[0] != "PARTICIPANT" {
		username = newUser.Account.Email
	}

	response := &api.UserAuthInfo{
		UserId:     id,
		Roles:      newUser.Roles,
		InstanceId: instanceID,
		Username:   username,
		ProfileId:  newUser.Profiles[0].ID.Hex(),
	}
	return response, nil
}

func (s *userManagementServer) CheckRefreshToken(ctx context.Context, req *api.RefreshTokenRequest) (*api.Status, error) {
	if req == nil || req.RefreshToken == "" || req.UserId == "" || req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	user, err := getUserByIDFromDB(req.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	err = user.RemoveRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "token not found")
	}
	user.ObjectInfos.LastTokenRefresh = time.Now().Unix()

	user, err = updateUserInDB(req.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "refresh token removed",
	}, nil
}

func (s *userManagementServer) TokenRefreshed(ctx context.Context, req *api.RefreshTokenRequest) (*api.Status, error) {
	if req == nil || req.RefreshToken == "" || req.UserId == "" || req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}
	user.AddRefreshToken(req.RefreshToken)
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
