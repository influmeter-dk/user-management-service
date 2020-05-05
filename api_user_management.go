package main

import (
	"context"
	"log"

	api "github.com/influenzanet/user-management-service/api"
	"github.com/influenzanet/user-management-service/models"
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

func (s *userManagementServer) ChangeAccountIDEmail(ctx context.Context, req *api.EmailChangeMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
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

	// TODO: send message to email
	log.Println("TODO: send email about successful account deletion")

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

func (s *userManagementServer) ChangePreferredLanguage(ctx context.Context, req *api.LanguageChangeMsg) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.LanguageCode == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	user, err := updateAccountPreferredLangDB(req.Token.InstanceId, req.Token.Id, req.LanguageCode)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) SaveProfile(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	if req.Profile.Id == "" {
		user.AddProfile(models.ProfileFromAPI(req.Profile))
	} else {
		err := user.UpdateProfile(models.ProfileFromAPI(req.Profile))
		if err != nil {
			return nil, status.Error(codes.Internal, "profile not found")
		}
	}

	updUser, err := updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return updUser.ToAPI(), nil
}

func (s *userManagementServer) RemoveProfile(ctx context.Context, req *api.ProfileRequest) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := getUserByIDFromDB(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	if len(user.Profiles) == 1 {
		return nil, status.Error(codes.Internal, "can't delete last profile")
	}

	if err := user.RemoveProfile(req.Profile.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	updUser, err := updateUserInDB(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return updUser.ToAPI(), nil
}

func (s *userManagementServer) UpdateContactPreferences(ctx context.Context, req *api.ContactPreferencesMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *userManagementServer) AddEmail(ctx context.Context, req *api.ContactInfoMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *userManagementServer) RemoveEmail(ctx context.Context, req *api.ContactInfoMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
