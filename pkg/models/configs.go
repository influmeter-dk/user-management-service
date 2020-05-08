package models

import "time"

type JWTConfig struct {
	TokenMinimumAgeMin  time.Duration // interpreted in minutes later
	TokenExpiryInterval time.Duration // interpreted in minutes later
}
