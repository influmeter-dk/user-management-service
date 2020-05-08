package tokens

import (
	"crypto/rand"
	b32 "encoding/base32"
	"strings"
	"time"
)

func GenerateUniqueTokenString() (string, error) {
	t := time.Now()
	ms := uint64(t.Unix())*1000 + uint64(t.Nanosecond()/int(time.Millisecond))

	token := make([]byte, 16)
	token[0] = byte(ms >> 40)
	token[1] = byte(ms >> 32)
	token[2] = byte(ms >> 24)
	token[3] = byte(ms >> 16)
	token[4] = byte(ms >> 8)
	token[5] = byte(ms)

	_, err := rand.Read(token[6:])
	if err != nil {
		return "", err
	}

	tokenStr := b32.StdEncoding.WithPadding(b32.NoPadding).EncodeToString(token)
	return tokenStr, nil
}

func GetExpirationTime(validityPeriod time.Duration) int64 {
	return time.Now().Add(validityPeriod).Unix()
}

func ReachedExpirationTime(t int64) bool {
	return time.Now().After(time.Unix(t, 0))
}

func GetRolesFromPayload(payload map[string]string) []string {
	roles := []string{}
	if val, ok := payload["roles"]; ok {
		roles = strings.Split(val, ",")
	}
	return roles
}

func GetUsernameFromPayload(payload map[string]string) string {
	username, ok := payload["username"]
	if !ok {
		return ""
	}
	return username
}
