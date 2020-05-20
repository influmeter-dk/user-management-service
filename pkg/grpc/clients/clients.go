package clients

import (
	"log"

	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
	"google.golang.org/grpc"
)

func connectToGRPCServer(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", addr, err)
	}
	return conn
}

func ConnectToMessagingService(addr string) (client messageAPI.MessagingServiceApiClient, close func() error) {
	// Connect to user management service
	serverConn := connectToGRPCServer(addr)
	return messageAPI.NewMessagingServiceApiClient(serverConn), serverConn.Close
}
