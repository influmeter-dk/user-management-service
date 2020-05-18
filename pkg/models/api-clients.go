package models

import (
	messageAPI "github.com/influenzanet/messaging-service/pkg/api/manage"
)

// APIClients holds the service clients to the internal services
type APIClients struct {
	MessagingService messageAPI.MessagingServiceApiClient
}
