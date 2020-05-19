package service

import (
	"context"
	"log"
	"time"

	messageAPI "github.com/influenzanet/messaging-service/pkg/api/manage"
	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/pwhash"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"github.com/influenzanet/user-management-service/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) CreateUser(ctx context.Context, req *api.CreateUserReq) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.AccountId == "" || req.InitialPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	if !utils.CheckRoleInToken(req.Token, "ADMIN") {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if !utils.CheckEmailFormat(req.AccountId) {
		return nil, status.Error(codes.InvalidArgument, "account id not a valid email")
	}
	if !utils.CheckPasswordFormat(req.InitialPassword) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password, err := pwhash.HashPassword(req.InitialPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create user DB object from request:
	newUser := models.User{
		Account: models.Account{
			Type:               "email",
			AccountID:          req.AccountId,
			AccountConfirmedAt: 0, // not confirmed yet
			Password:           password,
			PreferredLanguage:  req.PreferredLanguage,
		},
		Roles: req.Roles,
		Profiles: []models.Profile{
			{
				ID:       primitive.NewObjectID(),
				Alias:    req.AccountId,
				AvatarID: "default",
			},
		},
	}
	newUser.AddNewEmail(req.AccountId, false)

	instanceID := req.Token.InstanceId
	id, err := s.userDBservice.AddUser(instanceID, newUser)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	newUser.ID, _ = primitive.ObjectIDFromHex(id)

	// TempToken for contact verification:
	tempTokenInfos := models.TempToken{
		UserID:     id,
		InstanceID: instanceID,
		Purpose:    "contact-verification",
		Info: map[string]string{
			"type":  "email",
			"email": newUser.Account.AccountID,
		},
		Expiration: tokens.GetExpirationTime(time.Hour * 24 * 30),
	}
	tempToken, err := s.globalDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// ---> Trigger message sending
	_, err = s.clients.MessagingService.SendInstantEmail(ctx, &messageAPI.SendEmailReq{
		To:          []string{newUser.Account.AccountID},
		MessageType: "registration",
		ContentInfos: map[string]string{
			"token": tempToken,
		},
		PreferredLanguage: newUser.Account.PreferredLanguage,
	})
	if err != nil {
		log.Printf("CreateUser: %s", err.Error())
	}
	// <---

	return newUser.ToAPI(), nil
}

func (s *userManagementServer) AddRoleForUser(ctx context.Context, req *api.RoleMsg) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.AccountId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	if !utils.CheckRoleInToken(req.Token, "ADMIN") {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	user, err := s.userDBservice.GetUserByEmail(req.Token.InstanceId, req.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := user.AddRole(req.Role); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user, err = s.userDBservice.UpdateUser(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) RemoveRoleForUser(ctx context.Context, req *api.RoleMsg) (*api.User, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.AccountId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	if !utils.CheckRoleInToken(req.Token, "ADMIN") {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	user, err := s.userDBservice.GetUserByEmail(req.Token.InstanceId, req.AccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := user.RemoveRole(req.Role); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user, err = s.userDBservice.UpdateUser(req.Token.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return user.ToAPI(), nil
}

func (s *userManagementServer) FindNonParticipantUsers(ctx context.Context, req *api.FindNonParticipantUsersMsg) (*api.UserListMsg, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) {
		return nil, status.Error(codes.InvalidArgument, "missing arguments")
	}
	if !utils.CheckRoleInToken(req.Token, "ADMIN") {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	users, err := s.userDBservice.FindNonParticipantUsers(req.Token.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := api.UserListMsg{
		Users: make([]*api.User, len(users)),
	}
	for i, u := range users {
		resp.Users[i] = u.ToAPI()
	}
	return &resp, nil
}

func (s *userManagementServer) StreamUsers(req *api.StreamUsersMsg, stream api.UserManagementApi_StreamUsersServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}
