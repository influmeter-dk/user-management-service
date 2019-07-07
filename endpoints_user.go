package main

import (
	"context"
	"log"

	api "github.com/influenzanet/user-management-service/api"
	utils "github.com/influenzanet/user-management-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) GetUser(ctx context.Context, req *api.UserReference) (*api.User, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if req.UserId == "" {
		req.UserId = parsedToken.Id
	}

	if parsedToken.Id != req.UserId { // Later can be overwritten
		log.Printf("not authorized GetUser(): %s tried to access %s", parsedToken.Id, req.UserId)
		return nil, status.Error(codes.PermissionDenied, "not authorized")
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *api.PasswordChangeMsg) (*api.Status, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if !utils.CheckPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "new password too weak")
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, parsedToken.Id)
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

	err = updateUserPasswordInDB(parsedToken.InstanceId, parsedToken.Id, newHashedPw)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Printf("user %s initiated password change", parsedToken.Id)

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
	if req == nil || req.Token == "" || req.Name == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, parsedToken.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.Account.Name = nameFromAPI(req.Name)
	user, err = updateUserInDB(parsedToken.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) DeleteAccount(ctx context.Context, req *api.UserReference) (*api.Status, error) {
	if req == nil || req.Token == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// TODO: check if user auth is from admin - to remove user by admin
	if parsedToken.Id != req.UserId {
		log.Printf("unauthorized request: user %s initiated account removal for user id %s", parsedToken.Id, req.UserId)
		return nil, status.Error(codes.PermissionDenied, "not authorized")
	}
	log.Printf("user %s initiated account removal for user id %s", parsedToken.Id, req.UserId)

	if err := deleteUserFromDB(parsedToken.InstanceId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// remove all TempTokens for the given user ID using auth-service
	if _, err := clients.authService.PurgeUserTempTokens(context.Background(), &api.TempTokenInfo{
		UserId:     parsedToken.Id,
		InstanceId: parsedToken.InstanceId,
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
	// TODO: Update updated at time as well
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) UpdateChildren(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	// TODO: Update updated at time as well
	return nil, status.Error(codes.Unimplemented, "not implemented")
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
	if req == nil || req.Token == "" || req.SubProfile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, parsedToken.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.AddSubProfile(subProfileFromAPI(req.SubProfile))
	user, err = updateUserInDB(parsedToken.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) EditSubprofile(ctx context.Context, req *api.SubProfileRequest) (*api.User, error) {
	if req == nil || req.Token == "" || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, parsedToken.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	if err := user.UpdateSubProfile(subProfileFromAPI(req.SubProfile)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = updateUserInDB(parsedToken.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) RemoveSubprofile(ctx context.Context, req *api.SubProfileRequest) (*api.User, error) {
	if req == nil || req.Token == "" || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	parsedToken, err := clients.authService.ValidateJWT(context.Background(), &api.JWTRequest{
		Token: req.Token,
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := getUserByIDFromDB(parsedToken.InstanceId, parsedToken.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	if err := user.RemoveSubProfile(req.SubProfile.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = updateUserInDB(parsedToken.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}
