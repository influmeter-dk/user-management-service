package models

import "time"

type DBConfig struct {
	URI             string
	DBNamePrefix    string
	Timeout         int
	MaxPoolSize     uint64
	IdleConnTimeout int
}

type JWTConfig struct {
	TokenExpiryInterval time.Duration // interpreted in minutes later
}
