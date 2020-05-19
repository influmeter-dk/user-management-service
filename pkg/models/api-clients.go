package models

import (
	messageAPI "github.com/influenzanet/messaging-service/pkg/api/messaging_service"
)

// APIClients holds the service clients to the internal services
type APIClients struct {
	MessagingService messageAPI.MessagingServiceApiClient
}
