package utils

import (
	"regexp"
	"strings"

	"github.com/influenzanet/user-management-service/pkg/api"
)

// CheckEmailFormat to check if input string is a correct email address
func CheckEmailFormat(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$")

	return re.MatchString(email)
}

// CheckPasswordFormat to check if password fulfills password rules
func CheckPasswordFormat(password string) bool {
	if len(password) < 8 {
		return false
	}

	var res = 0

	lowercase := regexp.MustCompile("[a-z]")
	uppercase := regexp.MustCompile("[A-Z]")
	number := regexp.MustCompile(`\d`) //"^(?:(?=.*[a-z])(?:(?=.*[A-Z])(?=.*[\\d\\W])|(?=.*\\W)(?=.*\d))|(?=.*\W)(?=.*[A-Z])(?=.*\d)).{8,}$")
	symbol := regexp.MustCompile(`\W`)

	if lowercase.MatchString(password) || uppercase.MatchString(password) {
		res++
	}
	if number.MatchString(password) || symbol.MatchString(password) {
		res++
	}
	return res == 2
}

// IsTokenEmpty check a token from api if it's empty
func IsTokenEmpty(t *api.TokenInfos) bool {
	if t == nil || t.Id == "" || t.InstanceId == "" {
		return true
	}
	return false
}

// CheckRoleInToken Check if role is present in the token
func CheckRoleInToken(t *api.TokenInfos, role string) bool {
	if t == nil {
		return false
	}
	if val, ok := t.Payload["roles"]; ok {
		roles := strings.Split(val, ",")
		for _, r := range roles {
			if r == role {
				return true
			}
		}
	}
	return false
}
