package config

const (
	ENV_VERIFICATION_CODE_LIFETIME = "VERIFICATION_CODE_LIFETIME"
	ENV_TOKEN_EXPIRATION_MIN       = "TOKEN_EXPIRATION_MIN"

	ENV_USE_NO_CURSOR_TIMEOUT = "USE_NO_CURSOR_TIMEOUT"
)

const (
	defaultVerificationCodeLifetime = 15 * 60 // for 2FA 6 digit code
	defaultTokenExpirationMin       = 55
)
