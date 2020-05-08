package service

import (
	"github.com/influenzanet/user-management-service/pkg/dbs/globaldb"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	"github.com/influenzanet/user-management-service/pkg/models"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

type userManagementServer struct {
	clients         *models.APIClients
	userDBservice   *userdb.UserDBService
	globalDBService *globaldb.GlobalDBService
	JWT             models.JWTConfig
}

// NewUserManagementServer creates a new service instance
/* func NewUserManagementServer(
	clients *models.APIClients,
	userDBservice *userdb.UserDBService,
	globalDBservice *globaldb.GlobalDBService,
	JWT struct {
		TokenMinimumAgeMin  time.Duration // interpreted in minutes later
		TokenExpiryInterval time.Duration // interpreted in minutes later
	},
) *api.UserManagementApiServer {
	return &userManagementServer{
		clients:         clients,
		userDBservice:   userDBservice,
		globalDBService: globalDBservice,
		JWT:             JWT,
	}
}
*/
