package service

import (
	"context"

	"github.com/influenzanet/user-management-service/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) CreateUser(ctx context.Context, req *api.CreateUserReq) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
	/*if req == nil || utils.IsTokenEmpty(req.Token) || req.ContactInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}*/
}

func (s *userManagementServer) AddRoleForUser(ctx context.Context, req *api.RoleMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
	/*if req == nil || utils.IsTokenEmpty(req.Token) || req.ContactInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}*/
}

func (s *userManagementServer) RemoveRoleForUser(ctx context.Context, req *api.RoleMsg) (*api.User, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
	/*if req == nil || utils.IsTokenEmpty(req.Token) || req.ContactInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}*/
}

func (s *userManagementServer) FindNonParticipantUsers(ctx context.Context, req *api.FindNonParticipantUsersMsg) (*api.UserListMsg, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
	/*if req == nil || utils.IsTokenEmpty(req.Token) || req.ContactInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}*/
}
