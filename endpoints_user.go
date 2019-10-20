package main

import (
	"context"
	"log"
	"time"

	api "github.com/influenzanet/user-management-service/api"
	utils "github.com/influenzanet/user-management-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) GetUser(ctx context.Context, req *api.UserReference) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if req.UserId == "" {
		req.UserId = req.Token.Id
	}

	if req.Token.Id != req.UserId { // Later can be overwritten
		log.Printf("not authorized GetUser(): %s tried to access %s", req.Token.Id, req.UserId)
		return nil, status.Error(codes.PermissionDenied, "not authorized")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *api.PasswordChangeMsg) (*api.Status, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !utils.CheckPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "new password too weak")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	match, err := utils.ComparePasswordWithHash(user.Account.Password, req.OldPassword)
	if err != nil || !match {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	newHashedPw, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = updateUserPasswordInDB(req.Token.InstanceId, req.Token.Id, newHashedPw)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Printf("user %s initiated password change", req.Token.Id)

	// TODO: initiate email notification for user about password update

	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "password changed",
	}, nil
}

func (s *userManagementServer) ChangeEmail(ctx context.Context, req *api.EmailChangeMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) UpdateName(ctx context.Context, req *api.NameUpdateRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Name == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.Account.Name = nameFromAPI(req.Name)
	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) DeleteAccount(ctx context.Context, req *api.UserReference) (*api.Status, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	// TODO: check if user auth is from admin - to remove user by admin
	if req.Token.Id != req.UserId {
		log.Printf("unauthorized request: user %s initiated account removal for user id %s", req.Token.Id, req.UserId)
		return nil, status.Error(codes.PermissionDenied, "not authorized")
	}
	log.Printf("user %s initiated account removal for user id %s", req.Token.Id, req.UserId)

	if err := deleteUserFromDB(req.Token.InstanceId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// remove all TempTokens for the given user ID using auth-service
	if _, err := clients.authService.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{
		UserId:     req.Token.Id,
		InstanceId: req.Token.InstanceId,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Printf("user account with id %s successfully removed", req.UserId)
	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "user deleted",
	}, nil
}

func (s *userManagementServer) UpdateBirthDate(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.Profile.BirthDateUpdatedAt = time.Now().Unix()
	user.Profile.BirthDay = req.Profile.BirthDay
	user.Profile.BirthMonth = req.Profile.BirthMonth
	user.Profile.BirthYear = req.Profile.BirthYear

	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) UpdateChildren(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Profile == nil || req.Profile.Children == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.Profile.ChildrenUpdatedAt = time.Now().Unix()
	user.Profile.Children = childrenFromAPI(req.Profile.Children)

	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	return user.ToAPI(), nil
}

/*
TODO: remove
func (s *userManagementServer) UpdateProfile(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	if req == nil || req.Auth == nil || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.Profile = profileFromAPI(req.Profile)
	user, err = updateUserInDB(req.Auth.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}*/

func (s *userManagementServer) AddSubprofile(ctx context.Context, req *api.SubProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.SubProfile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.AddSubProfile(subProfileFromAPI(req.SubProfile))
	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) EditSubprofile(ctx context.Context, req *api.SubProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	if err := user.UpdateSubProfile(subProfileFromAPI(req.SubProfile)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) RemoveSubprofile(ctx context.Context, req *api.SubProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	if err := user.RemoveSubProfile(req.SubProfile.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}
