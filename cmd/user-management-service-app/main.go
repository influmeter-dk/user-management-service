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
	"github.com/influenzanet/user-management-service/pkg/timer_event"
)

func main() {
	conf := config.InitConfig()

	clients := &models.APIClients{}

	messagingClient, close := gc.ConnectToMessagingService(conf.ServiceURLs.MessagingService)
	defer close()
	clients.MessagingService = messagingClient

	loggingClient, close := gc.ConnectToLoggingService(conf.ServiceURLs.LoggingService)
	defer close()
	clients.LoggingService = loggingClient

	userDBService := userdb.NewUserDBService(conf.UserDBConfig)
	globalDBService := globaldb.NewGlobalDBService(conf.GlobalDBConfig)

	// Start timer thread
	userTimerService := timer_event.NewUserManagmentTimerService(
		60,
		globalDBService,
		userDBService,
		clients,
	)
	userTimerService.Run()

	// Start server thread
	ctx := context.Background()

	if err := service.RunServer(
		ctx,
		conf.Port,
		clients,
		userDBService,
		globalDBService,
		conf.JWT,
		conf.NewUserCountLimit,
	); err != nil {
		log.Fatal(err)
	}
}
