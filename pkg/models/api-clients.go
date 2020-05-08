package models

import "github.com/influenzanet/api-gateway/api"

// APIClients holds the service clients to the internal services
type APIClients struct {
	AuthService api.AuthServiceApiClient
}
