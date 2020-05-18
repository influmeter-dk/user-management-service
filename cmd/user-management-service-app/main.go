package main

import (
	"context"
	"log"

	"github.com/influenzanet/user-management-service/internal/config"
	"github.com/influenzanet/user-management-service/pkg/dbs/globaldb"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	"github.com/influenzanet/user-management-service/pkg/grpc/service"
	"github.com/influenzanet/user-management-service/pkg/models"
)

func main() {
	conf := config.InitConfig()

	clients := &models.APIClients{}
	// Connect to authentication service
	// authenticationServerConn := connectToAuthService()
	// defer authenticationServerConn.Close()
	// clients.authService = api.NewAuthServiceApiClient(authenticationServerConn)
	/*
	 */

	userDBService := userdb.NewUserDBService(conf.UserDBConfig)
	globalDBService := globaldb.NewGlobalDBService(conf.GlobalDBConfig)

	ctx := context.Background()

	if err := service.RunServer(
		ctx,
		conf.Port,
		clients,
		userDBService,
		globalDBService,
		conf.JWT,
	); err != nil {
		log.Fatal(err)
	}
}
