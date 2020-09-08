package service

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/pwhash"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"github.com/influenzanet/user-management-service/pkg/utils"

	constants "github.com/influenzanet/go-utils/pkg/constants"
)

func (s *userManagementServer) Status(ctx context.Context, _ *empty.Empty) (*api.ServiceStatus, error) {
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "service running",
		Version: apiVersion,
	}, nil
}

func (s *userManagementServer) SendVerificationCode(ctx context.Context, req *api.SendVerificationCodeReq) (*api.ServiceStatus, error) {
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if req.InstanceId == "" {
		req.InstanceId = "default"
	}

	req.Email = utils.SanitizeEmail(req.Email)
	user, err := s.userDBservice.GetUserByAccountID(req.InstanceId, req.Email)
	if err != nil {
		log.Printf("SECURITY WARNING: login step 1 attempt with wrong email address for %s", req.Email)
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if utils.HasMoreAttemptsRecently(user.Account.FailedLoginAttempts, allowedPasswordAttempts, loginFailedAttemptWindow) {
		log.Printf("SECURITY WARNING: login attempt blocked for email address for %s - too many wrong tries recently", req.Email)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if user.Account.VerificationCode.CreatedAt > time.Now().Unix()-loginVerificationCodeCooldown {
		return nil, status.Error(codes.InvalidArgument, "cannot generate verification code so often")
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		log.Printf("SECURITY WARNING: login step 1 attempt with wrong password for %s", user.ID.Hex())
		if err2 := s.userDBservice.SaveFailedLoginAttempt(req.InstanceId, user.ID.Hex()); err != nil {
			log.Printf("DB ERROR: unexpected error when updating user: %s ", err2.Error())
		}
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	err = s.generateAndSendVerificationCode(req.InstanceId, user)
	if err != nil {
		return nil, err
	}

	return &api.ServiceStatus{
		Version: apiVersion,
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "code generated and message sending triggered",
	}, nil
}

func (s *userManagementServer) AutoValidateTempToken(ctx context.Context, req *api.AutoValidateReq) (*api.AutoValidateResponse, error) {
	if req == nil || req.TempToken == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	tempToken, err := s.globalDBService.GetTempToken(req.TempToken)
	if err != nil {
		log.Printf("SECURITY WARNING: temptoken cannot be found %s", req.TempToken)
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	if tempToken.Purpose != constants.TOKEN_PURPOSE_SURVEY_LOGIN {
		log.Printf("SECURITY WARNING: temptoken with wrong prupose found: %s - by user %s in instance %s", tempToken.Purpose, tempToken.UserID, tempToken.InstanceID)
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	if tempToken.Expiration < time.Now().Unix() {
		log.Printf("SECURITY WARNING: temptoken is expired - by user %s in instance %s", tempToken.UserID, tempToken.InstanceID)
		return nil, status.Error(codes.InvalidArgument, "token expired")
	}

	user, err := s.userDBservice.GetUserByID(tempToken.InstanceID, tempToken.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}

	vc, err := tokens.GenerateVerificationCode(6)
	if err != nil {
		log.Printf("unexpected error while generating verification code: %v", err)
		return nil, status.Error(codes.Internal, "error while generating verification code")
	}

	user.Account.VerificationCode = models.VerificationCode{
		Code:      vc,
		ExpiresAt: time.Now().Unix() + verificationCodeLifetime,
	}
	user, err = s.userDBservice.UpdateUser(tempToken.InstanceID, user)
	if err != nil {
		log.Printf("AutoValidateTempToken: unexpected error when saving user -> %v", err)
		return nil, status.Error(codes.Internal, "user couldn't be updated")
	}

	sameUser := false
	if len(req.AccessToken) > 0 {
		validatedToken, _, err := tokens.ValidateToken(req.AccessToken)
		if err != nil {
			log.Printf("AutoValidateTempToken: unexpected error when parsing token -> %v", err)
		}
		if validatedToken.ID == tempToken.UserID && validatedToken.InstanceID == tempToken.InstanceID {
			sameUser = true
		}
	}

	return &api.AutoValidateResponse{AccountId: user.Account.AccountID, IsSameUser: sameUser, VerificationCode: vc, InstanceId: tempToken.InstanceID}, nil
}

func (s *userManagementServer) LoginWithEmail(ctx context.Context, req *api.LoginWithEmailMsg) (*api.LoginResponse, error) {
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if req.InstanceId == "" {
		req.InstanceId = "default"
	}

	req.Email = utils.SanitizeEmail(req.Email)
	user, err := s.userDBservice.GetUserByAccountID(req.InstanceId, req.Email)
	if err != nil {
		log.Printf("SECURITY WARNING: login attempt with wrong email address for %s", req.Email)
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if utils.HasMoreAttemptsRecently(user.Account.FailedLoginAttempts, allowedPasswordAttempts, loginFailedAttemptWindow) {
		log.Printf("SECURITY WARNING: login attempt blocked for email address for %s - too many wrong tries recently", req.Email)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		log.Printf("SECURITY WARNING: login attempt with wrong password for %s", user.ID.Hex())
		if err2 := s.userDBservice.SaveFailedLoginAttempt(req.InstanceId, user.ID.Hex()); err != nil {
			log.Printf("DB ERROR: unexpected error when updating user: %s ", err2.Error())
		}
		return nil, status.Error(codes.InvalidArgument, "invalid username and/or password")
	}

	if user.Account.AuthType == "2FA" {
		if user.Account.VerificationCode.Code == "" || req.VerificationCode == "" {
			err = s.generateAndSendVerificationCode(req.InstanceId, user)
			if err != nil {
				return nil, err
			}
			return &api.LoginResponse{
				User: &api.User{
					Account: &api.User_Account{
						AccountConfirmedAt: user.Account.AccountConfirmedAt,
						AccountId:          user.Account.AccountID,
					},
				},
				SecondFactorNeeded: true,
			}, nil
		}
		if user.Account.VerificationCode.ExpiresAt < time.Now().Unix() || user.Account.VerificationCode.Code != req.VerificationCode {
			log.Printf("SECURITY WARNING: login attempt with wrong or expired verification code for %s", user.ID.Hex())
			user.Account.VerificationCode = models.VerificationCode{}
			user, err = s.userDBservice.UpdateUser(req.InstanceId, user)
			if err != nil {
				log.Printf("LoginWithEmail: unexpected error when saving user -> %v", err)
			}

			if err2 := s.userDBservice.SaveFailedLoginAttempt(req.InstanceId, user.ID.Hex()); err != nil {
				log.Printf("DB ERROR: unexpected error when updating user: %s ", err2.Error())
			}
			return nil, status.Error(codes.InvalidArgument, "wrong verficiation code")
		}
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
		apiUser.Account.AccountConfirmedAt > 0,
		apiUser.Profiles[0].Id,
		currentRoles,
		req.InstanceId,
		s.JWT.TokenExpiryInterval,
		username,
		nil,
		otherProfileIDs,
	)
	if err != nil {
		log.Printf("LoginWithEmail: unexpected error during token generation -> %v", err)
		return nil, status.Error(codes.Internal, "token generation error")
	}

	// Refresh Token
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		log.Printf("LoginWithEmail: unexpected error during refresh token generation -> %v", err)
		return nil, status.Error(codes.Internal, "token generation error")
	}
	user.AddRefreshToken(rt)
	user.Timestamps.LastLogin = time.Now().Unix()
	user.Account.VerificationCode = models.VerificationCode{}
	user.Account.FailedLoginAttempts = utils.RemoveAttemptsOlderThan(user.Account.FailedLoginAttempts, 3600)
	user.Account.PasswordResetTriggers = utils.RemoveAttemptsOlderThan(user.Account.PasswordResetTriggers, 7200)

	user, err = s.userDBservice.UpdateUser(req.InstanceId, user)
	if err != nil {
		log.Printf("LoginWithEmail: unexpected error when saving user -> %v", err)
		return nil, status.Error(codes.Internal, "user couldn't be updated")
	}

	// remove all temptokens for password reset:
	if err := s.globalDBService.DeleteAllTempTokenForUser(req.InstanceId, user.ID.Hex(), constants.TOKEN_PURPOSE_PASSWORD_RESET); err != nil {
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

func (s *userManagementServer) SignupWithEmail(ctx context.Context, req *api.SignupWithEmailMsg) (*api.TokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	req.Email = utils.SanitizeEmail(req.Email)
	if !utils.CheckEmailFormat(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "email not valid")
	}
	if !utils.CheckLanguageCode(req.PreferredLanguage) {
		return nil, status.Error(codes.InvalidArgument, "language code wrong")
	}
	if !utils.CheckPasswordFormat(req.Password) {
		return nil, status.Error(codes.InvalidArgument, "password too weak")
	}

	if req.InstanceId == "" {
		req.InstanceId = "default"
	}

	newUserCount, err := s.userDBservice.CountRecentlyCreatedUsers(req.InstanceId, signupRateLimitWindow)
	if err != nil {
		log.Printf("ERROR: signup - unexpected error when counting: %v", err)
	} else {
		if newUserCount > s.newUserCountLimit {
			log.Println("ERROR: user creation blocked due to too many registations")
			return nil, status.Error(codes.Internal, "user creation failed, please try in some minutes again")
		}
	}

	password, err := pwhash.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create user DB object from request:
	newUser := models.User{
		Account: models.Account{
			Type:                  "email",
			AccountID:             req.Email,
			AccountConfirmedAt:    0, // not confirmed yet
			Password:              password,
			PreferredLanguage:     req.PreferredLanguage,
			FailedLoginAttempts:   []int64{},
			PasswordResetTriggers: []int64{},
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
	if req.Use_2Fa {
		newUser.Account.AuthType = "2FA"
	}

	if req.WantsNewsletter {
		newUser.ContactPreferences.SubscribedToNewsletter = true
		newUser.ContactPreferences.SendNewsletterTo = []string{newUser.ContactInfos[0].ID.Hex()}
	}
	// on which weekday the user will receive the reminder emails
	newUser.ContactPreferences.SubscribedToWeekly = true
	newUser.ContactPreferences.ReceiveWeeklyMessageDayOfWeek = int32(rand.Intn(7))

	id, err := s.userDBservice.AddUser(req.InstanceId, newUser)
	if err != nil {
		log.Printf("ERROR: when creating new user: %s", err.Error())
		return nil, status.Error(codes.Internal, "user creation failed")
	}
	newUser.ID, _ = primitive.ObjectIDFromHex(id)

	// TempToken for contact verification:
	tempTokenInfos := models.TempToken{
		UserID:     id,
		InstanceID: req.InstanceId,
		Purpose:    constants.TOKEN_PURPOSE_CONTACT_VERIFICATION,
		Info: map[string]string{
			"type":  "email",
			"email": newUser.Account.AccountID,
		},
		Expiration: tokens.GetExpirationTime(time.Hour * 24 * 30),
	}
	tempToken, err := s.globalDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		log.Printf("ERROR: signup method failed to create verification token: %s", err.Error())
		return nil, status.Error(codes.Internal, "failed to create verification token")
	}

	// ---> Trigger message sending
	go func(instanceID string, accountID string, tempToken string, preferredLang string) {
		_, err = s.clients.MessagingService.SendInstantEmail(context.TODO(), &messageAPI.SendEmailReq{
			InstanceId:  instanceID,
			To:          []string{accountID},
			MessageType: constants.EMAIL_TYPE_REGISTRATION,
			ContentInfos: map[string]string{
				"token": tempToken,
			},
			PreferredLanguage: preferredLang,
		})
		if err != nil {
			log.Printf("SignupWithEmail: %s", err.Error())
		}
	}(req.InstanceId, newUser.Account.AccountID, tempToken, newUser.Account.PreferredLanguage)
	// <---

	var username string
	if len(newUser.Roles) > 1 || len(newUser.Roles) == 1 && newUser.Roles[0] != "PARTICIPANT" {
		username = newUser.Account.AccountID
	}
	apiUser := newUser.ToAPI()

	// Access Token
	token, err := tokens.GenerateNewToken(
		apiUser.Id,
		apiUser.Account.AccountConfirmedAt > 0,
		apiUser.Profiles[0].Id,
		newUser.Roles,
		req.InstanceId,
		s.JWT.TokenExpiryInterval,
		username,
		nil,
		[]string{},
	)
	if err != nil {
		log.Printf("ERROR: signup method failed to generate jwt: %s", err.Error())
		return nil, status.Error(codes.Internal, "token creation failed")
	}

	// Refresh Token
	rt, err := tokens.GenerateUniqueTokenString()
	if err != nil {
		log.Printf("ERROR: signup method failed to generate refresh token: %s", err.Error())
		return nil, status.Error(codes.Internal, "token creation failed")
	}
	newUser.AddRefreshToken(rt)
	newUser.Timestamps.LastLogin = time.Now().Unix()

	newUser, err = s.userDBservice.UpdateUser(req.InstanceId, newUser)
	if err != nil {
		log.Printf("ERROR: signup method failed to save refresh token: %s", err.Error())
		return nil, status.Error(codes.Internal, "user created, but token could not be saved")
	}

	s.SaveLogEvent(req.InstanceId, newUser.ID.Hex(), loggingAPI.LogEventType_LOG, "account created", newUser.Account.AccountID)

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
			log.Printf("ERROR: SwitchProfile method failed to save user: %s", err.Error())
			return nil, status.Error(codes.Internal, "user couldn't be updated")
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
		apiUser.Account.AccountConfirmedAt > 0,
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

	tokenInfos, err := s.ValidateTempToken(req.Token, []string{
		constants.TOKEN_PURPOSE_CONTACT_VERIFICATION,
		constants.TOKEN_PURPOSE_INVITATION,
	})
	if err != nil {
		log.Printf("VerifyContact: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.userDBservice.GetUserByID(tokenInfos.InstanceID, tokenInfos.UserID)
	if err != nil {
		log.Printf("VerifyContact: %s", err.Error())
		return nil, status.Error(codes.InvalidArgument, "no user found")
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

	if tokenInfos.Purpose != constants.TOKEN_PURPOSE_INVITATION {
		// invitation token will be required for password reset, so we don't remove it
		if err := s.globalDBService.DeleteTempToken(req.Token); err != nil {
			log.Printf("VerifyContact delete token: %s", err.Error())
		}
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

	if ci.ConfirmationLinkSentAt > time.Now().Unix()-contactVerificationMessageCooldown {
		return nil, status.Error(codes.InvalidArgument, "cannot send verification so often")
	}

	// TempToken for contact verification:
	tempTokenInfos := models.TempToken{
		UserID:     req.Token.Id,
		InstanceID: req.Token.InstanceId,
		Purpose:    constants.TOKEN_PURPOSE_CONTACT_VERIFICATION,
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
		MessageType: constants.EMAIL_TYPE_VERIFY_EMAIL,
		ContentInfos: map[string]string{
			"token": tempToken,
		},
		PreferredLanguage: user.Account.PreferredLanguage,
	})
	if err != nil {
		log.Printf("ResendContactVerification: %s", err.Error())
	}
	// <---

	// update last verification email sent time:
	user.SetContactInfoVerificationSent("email", req.Address)
	_, err = s.userDBservice.UpdateUser(req.Token.InstanceId, user)
	if err != nil {
		log.Printf("ResendContactVerification: %s", err.Error())
	}

	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "message sent",
		Version: apiVersion,
	}, nil
}
