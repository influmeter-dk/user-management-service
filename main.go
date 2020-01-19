package main

import (
	"log"
	"net"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"

	api "github.com/influenzanet/user-management-service/api"
)

type userManagementServer struct {
}

// APIClients holds the service clients to the internal services
type APIClients struct {
	authService api.AuthServiceApiClient
}

var clients = APIClients{}

var dbClient *mongo.Client
var conf Config

func connectToAuthService() *grpc.ClientConn {
	conn, err := grpc.Dial(conf.ServiceURLs.AuthService, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return conn
}

func init() {
	initConfig()
	dbInit()
}

func main() {
	// Connect to authentication service
	authenticationServerConn := connectToAuthService()
	defer authenticationServerConn.Close()
	clients.authService = api.NewAuthServiceApiClient(authenticationServerConn)

	lis, err := net.Listen("tcp", ":"+conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("wait connections on port " + conf.Port)

	grpcServer := grpc.NewServer()
	api.RegisterUserManagementApiServer(grpcServer, &userManagementServer{})
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
