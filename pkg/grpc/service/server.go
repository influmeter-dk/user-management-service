package service

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/dbs/globaldb"
	"github.com/influenzanet/user-management-service/pkg/dbs/userdb"
	"github.com/influenzanet/user-management-service/pkg/models"
	"google.golang.org/grpc"
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
func NewUserManagementServer(
	clients *models.APIClients,
	userDBservice *userdb.UserDBService,
	globalDBservice *globaldb.GlobalDBService,
	JWT struct {
		TokenExpiryInterval time.Duration // interpreted in minutes later
	},
) api.UserManagementApiServer {
	return &userManagementServer{
		clients:         clients,
		userDBservice:   userDBservice,
		globalDBService: globalDBservice,
		JWT:             JWT,
	}
}

// RunServer runs gRPC service to publish ToDo service
func RunServer(ctx context.Context, port string,
	clients *models.APIClients,
	userDBservice *userdb.UserDBService,
	globalDBservice *globaldb.GlobalDBService,
	JWT struct {
		TokenExpiryInterval time.Duration // interpreted in minutes later
	},
) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// register service
	server := grpc.NewServer()
	api.RegisterUserManagementApiServer(server, NewUserManagementServer(
		clients,
		userDBservice,
		globalDBservice,
		JWT,
	))

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Println("starting gRPC server...")
	log.Println("wait connections on port " + port)
	return server.Serve(lis)
}
