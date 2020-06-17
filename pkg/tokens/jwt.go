package tokens

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	b64 "encoding/base64"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/influenzanet/user-management-service/pkg/models"
)

var (
	secretKey    []byte
	secretKeyEnc string
)

// UserClaims - Information a token enocodes
type UserClaims struct {
	ID               string            `json:"id,omitempty"`
	InstanceID       string            `json:"instance_id,omitempty"`
	ProfileID        string            `json:"profile_id,omitempty"`
	Payload          map[string]string `json:"payload,omitempty"`
	AccountConfirmed bool              `json:"accountConfirmed,omitempty"`
	TempTokenInfos   *models.TempToken `json:"temptoken,omitempty"`
	OtherProfileIDs  []string          `json:"other_profile_ids,omitempty"`
	jwt.StandardClaims
}

func getSecretKey() (newSecretKey []byte, err error) {
	newSecretKeyEnc := os.Getenv("JWT_TOKEN_KEY")
	if secretKeyEnc == newSecretKeyEnc {
		return newSecretKey, nil
	}
	secretKeyEnc = newSecretKeyEnc
	newSecretKey, err = b64.StdEncoding.DecodeString(newSecretKeyEnc)
	if err != nil {
		return newSecretKey, err
	}
	if len(newSecretKey) < 32 {
		return newSecretKey, errors.New("couldn't find proper secret key")
	}
	secretKey = newSecretKey
	return
}

// GenerateNewToken create and signes a new token
func GenerateNewToken(userID string, accountConfirmed bool, profileID string, userRoles []string, instanceID string, experiresIn time.Duration, username string, tempTokenInfos *models.TempToken, otherProfileIDs []string) (string, error) {
	payload := map[string]string{}

	if len(userRoles) > 0 {
		payload["roles"] = strings.Join(userRoles, ",")
	}
	if len(username) > 0 {
		payload["username"] = username
	}

	// Create the Claims
	claims := UserClaims{
		userID,
		instanceID,
		profileID,
		payload,
		accountConfirmed,
		tempTokenInfos,
		otherProfileIDs,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(experiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	_, err := getSecretKey()
	if err != nil {
		return "", err
	}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secretKey)
	return tokenString, err
}

// ValidateToken parses and validates the token string
func ValidateToken(tokenString string) (claims *UserClaims, valid bool, err error) {
	_, err = getSecretKey()
	if err != nil {
		return nil, false, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if token == nil {
		return
	}
	claims, valid = token.Claims.(*UserClaims)
	valid = valid && token.Valid
	return
}
