package main

import (
	"context"
	"log"
	"regexp"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	influenzanet "github.com/influenzanet/api/dist/go"
	user_api "github.com/influenzanet/api/dist/go/user-management"
)

func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

func comparePasswordWithHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func checkEmailFormat(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	return re.MatchString(email)
}

func checkPasswordFormat(password string) bool {
	if len(password) < 8 {
		return false
	}

	var res = 0

	lowercase := regexp.MustCompile("[a-z]")
	uppercase := regexp.MustCompile("[A-Z]")
	number := regexp.MustCompile("\\d") //"^(?:(?=.*[a-z])(?:(?=.*[A-Z])(?=.*[\\d\\W])|(?=.*\\W)(?=.*\d))|(?=.*\W)(?=.*[A-Z])(?=.*\d)).{8,}$")
	symbol := regexp.MustCompile("\\W")

	if lowercase.MatchString(password) {
		res++
	}
	if uppercase.MatchString(password) {
		res++
	}
	if number.MatchString(password) {
		res++
	}
	if symbol.MatchString(password) {
		res++
	}

	return res >= 3
}

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

func (s *userManagementServer) ChangePassword(ctx context.Context, req *user_api.PasswordChangeMsg) (*influenzanet.Status, error) {
	if req == nil || req.Auth == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	if !checkPasswordFormat(req.NewPassword) {
		return nil, status.Error(codes.InvalidArgument, "new password too weak")
	}

	user, err := findUserByID(req.Auth.InstanceId, req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	if comparePasswordWithHash(user.Password, req.OldPassword) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user and/or password")
	}

	newHashedPw := hashPassword(req.NewPassword)
	err = updateUserPasswordDB(req.Auth.InstanceId, req.Auth.UserId, newHashedPw)
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
