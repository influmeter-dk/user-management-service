package main

import (
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

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
}

func main() {
	// Connect to authentication service
	// authenticationServerConn := connectToAuthService()
	// defer authenticationServerConn.Close()
	// clients.authService = api.NewAuthServiceApiClient(authenticationServerConn)
	/*
	 */
}
