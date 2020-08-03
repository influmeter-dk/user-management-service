package models

import (
	loggingAPI "github.com/influenzanet/logging-service/pkg/api"
	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
)

// APIClients holds the service clients to the internal services
type APIClients struct {
	MessagingService messageAPI.MessagingServiceApiClient
	LoggingService   loggingAPI.LoggingServiceApiClient
}
