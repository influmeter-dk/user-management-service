package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
var conf config

func dbInit() {
	dbCreds, err := readDBcredentials(conf.DB.CredentialsPath)
	if err != nil {
		log.Fatal(err)
	}

	// mongodb+srv://user-management-service:<PASSWORD>@influenzanettestdbcluster-pwvbz.mongodb.net/test?retryWrites=true
	address := fmt.Sprintf(`mongodb+srv://%s:%s@%s`, dbCreds.Username, dbCreds.Password, conf.DB.Address)

	dbClient, err = mongo.NewClient(options.Client().ApplyURI(address))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func connectToAuthService() *grpc.ClientConn {
	conn, err := grpc.Dial(conf.ServiceURLs.AuthService, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return conn
}

func init() {
	readConfig()
	dbInit()
}

func main() {
	// Connect to authentication service
	authenticationServerConn := connectToAuthService()
	defer authenticationServerConn.Close()
	clients.authService = api.NewAuthServiceApiClient(authenticationServerConn)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	log.Println("wait connections on port " + strconv.Itoa(conf.Port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterUserManagementApiServer(grpcServer, &userManagementServer{})
	grpcServer.Serve(lis)
}
