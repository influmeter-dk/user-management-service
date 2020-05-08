package service

import (
	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/userdb"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

type userManagementServer struct {
	clients       *models.APIClients
	userDBservice *userdb.UserDBService
}

// NewUserManagementServer creates a new service instance
func NewUserManagementServer(
	clients *models.APIClients,
	userDBservice *userdb.UserDBService,
) *api.UserManagementApiServer {
	return &userManagementServer{
		clients:       clients,
		userDBservice: userDBservice,
	}
}
