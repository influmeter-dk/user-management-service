package service

const (
	verificationCodeLifetime           = 5 * 60 // for 2FA 6 digit code
	contactVerificationMessageCooldown = 1 * 60
	loginVerificationCodeCooldown      = 20

	signupRateLimitWindow      = 5 * 60
	loginFailedAttemptWindow   = 5 * 50  // seconds
	passwordResetAttemptWindow = 60 * 60 // 1 hour
)
