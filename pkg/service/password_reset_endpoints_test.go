package service

import "testing"

func TestInitiatePasswordResetEndpoint(t *testing.T) {
	// create test user

	// with missing
	// with empty
	// with wrong account id
	// with valid account id
	t.Error("test unimplemented")
}

func TestGetInfosForPasswordResetEndpoint(t *testing.T) {
	// create test user
	// generate temptoken for password reset

	// with missing
	// with empty
	// with wrong token
	// with expired token
	// with wrong token purpose
	// with valid token
	t.Error("test unimplemented")
}

func TestResetPasswordEndpoint(t *testing.T) {
	// create test user
	// generate temptoken for password reset

	// with missing
	// with empty
	// with wrong token
	// with expired token
	// with wrong token purpose
	// with too weak password
	// with valid agruments
	t.Error("test unimplemented")
}
