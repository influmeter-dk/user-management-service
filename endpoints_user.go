package main

import (
	"context"
	"log"

	influenzanet "github.com/influenzanet/api/dist/go"
	user_api "github.com/influenzanet/api/dist/go/user-management"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) GetUser(ctx context.Context, req *user_api.UserReference) (*user_api.User, error) {
	if req == nil || req.Auth == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if req.Auth.UserId != req.UserId { // Later can be overwritten
		log.Printf("not authorized: %s tried to access %s", req.Auth.UserId, req.UserId)
		return nil, status.Error(codes.PermissionDenied, "not authorized")
	}

	user, err := getUserByIDFromDB(req.Auth.InstanceId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) ChangePassword(ctx context.Context, req *user_api.PasswordChangeMsg) (*influenzanet.Status, error) {
	if req == nil || req.Auth == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !checkPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "new password too weak")
	}

	user, err := getUserByIDFromDB(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	if comparePasswordWithHash(user.Account.Password, req.OldPassword) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	newHashedPw := hashPassword(req.NewPassword)
	err = updateUserPasswordInDB(req.Auth.InstanceId, req.Auth.UserId, newHashedPw)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Printf("user %s initiated password change", req.Auth.UserId)

	// TODO: initiate email notification for user about password update

	return &influenzanet.Status{
		Status: influenzanet.Status_NORMAL,
		Msg:    "password changed",
	}, nil
}

func (s *userManagementServer) ChangeEmail(ctx context.Context, req *user_api.EmailChangeMsg) (*user_api.User, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) UpdateName(ctx context.Context, req *user_api.NameUpdateRequest) (*user_api.User, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) UpdateBirthDate(ctx context.Context, req *user_api.ProfileRequest) (*user_api.User, error) {
	// TODO: Update updated at time as well
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *userManagementServer) UpdateChildren(ctx context.Context, req *user_api.ProfileRequest) (*user_api.User, error) {
	// TODO: Update updated at time as well
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

/*
TODO: remove
func (s *userManagementServer) UpdateProfile(ctx context.Context, req *user_api.ProfileRequest) (*user_api.User, error) {
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

func (s *userManagementServer) AddSubprofile(ctx context.Context, req *user_api.SubProfileRequest) (*user_api.User, error) {
	if req == nil || req.Auth == nil || req.SubProfile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	user, err := getUserByIDFromDB(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	user.AddSubProfile(subProfileFromAPI(req.SubProfile))
	user, err = updateUserInDB(req.Auth.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) EditSubprofile(ctx context.Context, req *user_api.SubProfileRequest) (*user_api.User, error) {
	if req == nil || req.Auth == nil || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	user, err := getUserByIDFromDB(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "not found")
	}

	if err := user.UpdateSubProfile(subProfileFromAPI(req.SubProfile)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err = updateUserInDB(req.Auth.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToAPI(), nil
}

func (s *userManagementServer) RemoveSubprofile(ctx context.Context, req *user_api.SubProfileRequest) (*user_api.User, error) {
	/*if req == nil || req.Auth == nil || req.SubProfile == nil || req.SubProfile.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}*/
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
