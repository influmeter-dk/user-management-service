package service

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/pwhash"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"github.com/influenzanet/user-management-service/pkg/utils"
)

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*api.ServiceStatus, error) {
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "service running",
		Version: apiVersion,
	}, nil
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, req *api.LoginWithEmailMsg) (*api.LoginResponse, error) {
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	instanceID := req.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}
	user, err := s.userDBservice.GetUserByAccountID(instanceID, req.Email)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	var username string

	currentRoles := user.Roles
	if req.AsParticipant {
		currentRoles = []string{"PARTICIPANT"}
	} else {
		if len(user.Roles) > 1 || len(user.Roles) == 1 && user.Roles[0] != "PARTICIPANT" {
			username = user.Account.AccountID
		}
	}

	apiUser := user.ToAPI()
	otherProfileIDs := []string{}
	for _, p := range apiUser.Profiles {
		if p.Id != apiUser.Profiles[0].Id {
			otherProfileIDs = append(otherProfileIDs, p.Id)
		}
	}
	// Access Token
	token, err := tokens.GenerateNewToken(
		apiUser.Id,
		apiUser.Profiles[0].Id,
		currentRoles,
		instanceID,
		s.JWT.TokenExpiryInterval,
		username,
		nil,
		otherProfileIDs,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Refresh Token
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user.AddRefreshToken(rt)
	user.Timestamps.LastLogin = time.Now().Unix()

	user, err = s.userDBservice.UpdateUser(req.InstanceId, user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// remove all temptokens for password reset:
	if err := s.globalDBService.DeleteAllTempTokenForUser(instanceID, user.ID.Hex(), "password-reset"); err != nil {
		log.Printf("LoginWithEmail: %s", err.Error())
	}

	response := &api.LoginResponse{
		Token: &api.TokenResponse{
			AccessToken:       token,
			RefreshToken:      rt,
			ExpiresIn:         int32(s.JWT.TokenExpiryInterval / time.Minute),
			Profiles:          apiUser.Profiles,
			SelectedProfileId: apiUser.Profiles[0].Id,
			PreferredLanguage: apiUser.Account.PreferredLanguage,
		},
		User: user.ToAPI(),
	}
	return response, nil

}

func (s *userManagementServer) LoginWithTempToken(ctx context.Context, req *api.JWTRequest) (*api.LoginResponse, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tokenInfos, err := s.globalDBService.GetTempToken(req.Token)
	if err != nil {
		log.Println(err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	if tokenInfos.Purpose != "survey-login" || tokens.ReachedExpirationTime(tokenInfos.Expiration) {
		log.Println("wrong token found for survey login:")
		log.Println(tokenInfos)
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	user, err := s.userDBservice.GetUserByID(tokenInfos.InstanceID, tokenInfos.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}

	currentRoles := []string{"PARTICIPANT"}

	apiUser := user.ToAPI()
	otherProfileIDs := []string{}
	for _, p := range apiUser.Profiles {
		if p.Id != apiUser.Profiles[0].Id {
			otherProfileIDs = append(otherProfileIDs, p.Id)
		}
	}
	// Access Token
	token, err := tokens.GenerateNewToken(
		apiUser.Id,
		apiUser.Profiles[0].Id,
		currentRoles,
		tokenInfos.InstanceID,
		s.JWT.TokenExpiryInterval,
		"",
		&tokenInfos,
		otherProfileIDs,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.userDBservice.UpdateLoginTime(tokenInfos.InstanceID, user.ID.Hex()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	apiUser.ContactInfos = []*api.ContactInfo{}
	apiUser.Timestamps = nil
	apiUser.ContactPreferences = nil
	apiUser.Roles = []string{}

	response := &api.LoginResponse{
		Token: &api.TokenResponse{
			AccessToken:       token,
			ExpiresIn:         int32(s.JWT.TokenExpiryInterval / time.Minute),
			Profiles:          apiUser.Profiles,
			SelectedProfileId: apiUser.Profiles[0].Id,
			PreferredLanguage: apiUser.Account.PreferredLanguage,
		},
		User: apiUser,
	}
	return response, nil
}

func (s *userManagementServer) SignupWithEmail(ctx context.Context, req *api.SignupWithEmailMsg) (*api.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if !utils.CheckEmailFormat(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "email not valid")
	}
	if !utils.CheckPasswordFormat(req.Password) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	password, err := pwhash.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create user DB object from request:
	newUser := models.User{
		Account: models.Account{
			Type:               "email",
			AccountID:          req.Email,
			AccountConfirmedAt: 0, // not confirmed yet
			Password:           password,
			PreferredLanguage:  req.PreferredLanguage,
		},
		Roles: []string{"PARTICIPANT"},
		Profiles: []models.Profile{
			{
				ID:                 primitive.NewObjectID(),
				Alias:              req.Email,
				ConsentConfirmedAt: time.Now().Unix(),
				AvatarID:           "default",
			},
		},
		Timestamps: models.Timestamps{
			CreatedAt: time.Now().Unix(),
		},
	}
	newUser.AddNewEmail(req.Email, false)

	if req.WantsNewsletter {
		newUser.ContactPreferences.SubscribedToNewsletter = true
		newUser.ContactPreferences.SendNewsletterTo = []string{newUser.ContactInfos[0].ID.Hex()}
	}

	instanceID := req.InstanceId
	if instanceID == "" {
		instanceID = "default"
	}

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
		InstanceId:  instanceID,
		To:          []string{newUser.Account.AccountID},
		MessageType: "registration",
		ContentInfos: map[string]string{
			"token": tempToken,
		},
		PreferredLanguage: newUser.Account.PreferredLanguage,
	})
	if err != nil {
		log.Printf("SignupWithEmail: %s", err.Error())
	}
	// <---

	var username string
	if len(newUser.Roles) > 1 || len(newUser.Roles) == 1 && newUser.Roles[0] != "PARTICIPANT" {
		username = newUser.Account.AccountID
	}
	apiUser := newUser.ToAPI()

	// Access Token
	token, err := tokens.GenerateNewToken(
		apiUser.Id,
		apiUser.Profiles[0].Id,
		newUser.Roles,
		instanceID,
		s.JWT.TokenExpiryInterval,
		username,
		nil,
		[]string{},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Refresh Token
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	newUser.AddRefreshToken(rt)
	newUser.Timestamps.LastLogin = time.Now().Unix()

	newUser, err = s.userDBservice.UpdateUser(req.InstanceId, newUser)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &api.TokenResponse{
		AccessToken:       token,
		RefreshToken:      rt,
		ExpiresIn:         int32(s.JWT.TokenExpiryInterval / time.Minute),
		Profiles:          apiUser.Profiles,
		SelectedProfileId: apiUser.Profiles[0].Id,
		PreferredLanguage: apiUser.Account.PreferredLanguage,
	}
	return response, nil
}

func (s *userManagementServer) SwitchProfile(ctx context.Context, req *api.SwitchProfileRequest) (*api.TokenResponse, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.ProfileId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if req.Token.TempToken == nil && req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := s.userDBservice.GetUserByID(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "user not found")
	}

	profile, err := user.FindProfile(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.Internal, "profile not found")
	}

	var rt string
	if req.RefreshToken != "" {
		// only if not temptoken
		if err := user.RemoveRefreshToken(req.RefreshToken); err != nil {
			return nil, status.Error(codes.Internal, "wrong refresh token")
		}
		// Refresh Token
		rt, err = tokens.GenerateUniqueTokenString()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		user.AddRefreshToken(rt)

		user, err = s.userDBservice.UpdateUser(req.Token.InstanceId, user)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	var username string
	if len(user.Roles) > 1 || len(user.Roles) == 1 && user.Roles[0] != "PARTICIPANT" {
		username = user.Account.AccountID
	}
	apiUser := user.ToAPI()
	otherProfileIDs := []string{}
	for _, p := range apiUser.Profiles {
		if p.Id != req.ProfileId {
			otherProfileIDs = append(otherProfileIDs, p.Id)
		}
	}

	// Access Token
	token, err := tokens.GenerateNewToken(
		apiUser.Id,
		profile.ID.Hex(),
		user.Roles,
		req.Token.InstanceId,
		s.JWT.TokenExpiryInterval,
		username,
		models.TempTokenFromAPI(req.Token.TempToken),
		otherProfileIDs,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &api.TokenResponse{
		AccessToken:       token,
		RefreshToken:      rt,
		ExpiresIn:         int32(s.JWT.TokenExpiryInterval / time.Minute),
		Profiles:          apiUser.Profiles,
		SelectedProfileId: profile.ID.Hex(),
		PreferredLanguage: apiUser.Account.PreferredLanguage,
	}
	return response, nil
}

func (s *userManagementServer) VerifyContact(ctx context.Context, req *api.TempToken) (*api.User, error) {
	if req == nil || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	tokenInfos, err := s.ValidateTempToken(req.Token, "contact-verification")
	if err != nil {
		log.Printf("VerifyContact: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.userDBservice.GetUserByID(tokenInfos.InstanceID, tokenInfos.UserID)
	if err != nil {
		log.Printf("VerifyContact: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cType, ok1 := tokenInfos.Info["type"]
	email, ok2 := tokenInfos.Info["email"]
	if !ok1 || !ok2 {
		return nil, status.Error(codes.InvalidArgument, "missing token info")
	}

	if err := user.ConfirmContactInfo(cType, email); err != nil {
		log.Printf("VerifyContact: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if user.Account.Type == "email" && user.Account.AccountID == email {
		user.Account.AccountConfirmedAt = time.Now().Unix()
	}
	user, err = s.userDBservice.UpdateUser(tokenInfos.InstanceID, user)

	if err := s.globalDBService.DeleteTempToken(req.Token); err != nil {
		log.Printf("VerifyContact delete token: %s", err.Error())
	}
	return user.ToAPI(), err
}

func (s *userManagementServer) ResendContactVerification(ctx context.Context, req *api.ResendContactVerificationReq) (*api.ServiceStatus, error) {
	if req == nil || utils.IsTokenEmpty(req.Token) || req.Address == "" || req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	user, err := s.userDBservice.GetUserByID(req.Token.InstanceId, req.Token.Id)
	if err != nil {
		log.Printf("ResendContactVerification: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ci, found := user.FindContactInfoByTypeAndAddr("email", req.Address)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "address not found")
	}

	// TempToken for contact verification:
	tempTokenInfos := models.TempToken{
		UserID:     req.Token.Id,
		InstanceID: req.Token.InstanceId,
		Purpose:    "contact-verification",
		Info: map[string]string{
			"type":  "email",
			"email": ci.Email,
		},
		Expiration: tokens.GetExpirationTime(time.Hour * 24 * 30),
	}
	tempToken, err := s.globalDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// ---> Trigger message sending
	_, err = s.clients.MessagingService.SendInstantEmail(ctx, &messageAPI.SendEmailReq{
		InstanceId:  req.Token.InstanceId,
		To:          []string{req.Address},
		MessageType: "verify-email",
		ContentInfos: map[string]string{
			"token": tempToken,
		},
		PreferredLanguage: user.Account.PreferredLanguage,
	})
	if err != nil {
		log.Printf("ResendContactVerification: %s", err.Error())
	}
	// <---

	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "message sent",
		Version: apiVersion,
	}, nil
}
