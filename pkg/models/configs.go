package models

import "time"

type DBConfig struct {
	URI             string
	DBNamePrefix    string
	Timeout         int
	NoCursorTimeout bool
	MaxPoolSize     uint64
	IdleConnTimeout int
}

type Intervals struct {
	TokenExpiryInterval      time.Duration // interpreted in minutes later
	VerificationCodeLifetime int64         // in seconds
}
