package main

import (
	"context"
	"log"

	"github.com/influenzanet/user-management-service/internal/config"
	"github.com/influenzanet/user-management-service/pkg/dbs/globaldb"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	gc "github.com/influenzanet/user-management-service/pkg/grpc/clients"
	"github.com/influenzanet/user-management-service/pkg/grpc/service"
	"github.com/influenzanet/user-management-service/pkg/models"
)

func main() {
	conf := config.InitConfig()

	clients := &models.APIClients{}

	messagingClient, close := gc.ConnectToMessagingSerive(conf.ServiceURLs.MessagingService)
	defer close()
	clients.MessagingService = messagingClient

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
