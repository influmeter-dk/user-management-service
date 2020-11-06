package service

const (
	contactVerificationMessageCooldown = 1 * 60
	loginVerificationCodeCooldown      = 20

	signupRateLimitWindow           = 5 * 60
	loginFailedAttemptWindow        = 5 * 50  // seconds
	passwordResetAttemptWindow      = 60 * 60 // 1 hour
	allowedPasswordAttempts         = 10
	allowedVerificationCodeAttempts = 3

	userCreationTimestampOffset = 7 * 24 * 3600 // consider user deletion only after this time, when created by admin

	maximumProfilesAllowed = 6
)
