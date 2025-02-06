package util

import (
	"regexp"
	"strings"
)

func ValidEmail(email string) bool {
	email = strings.TrimSpace(email) // Trim spaces
	pattern := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

func ValidHashpass(hashpass string) bool {
    if len(hashpass) != 64 {
        return false
    }
	// Validate a 64-character alphanumeric 'hashpass'.
    match, _ := regexp.MatchString("^[a-f0-9]{64}$", hashpass)
    return match
}

func ValidOTP(otp string) bool {
	otp = strings.TrimSpace(otp)
	pattern := `^[A-Za-z0-9]{6}$` // Matches exactly 6 alphanumeric characters
	re := regexp.MustCompile(pattern)
	return re.MatchString(otp)
}

func ValidToken(token string) (bool) {
	// Split the token into uid and 'hashpass'.
	parts := strings.Split(strings.ToLower(token), "::")
	if len(parts) != 2 {
		return false
	}
	uid, hashpass := parts[0], parts[1]

	// Validate the uid (assuming it's an email) and check if 'hashpass' is a 64-char lowercase alphanumeric string.
	if !ValidEmail(uid) || !ValidHashpass(hashpass) {
		return false
	}

	return true
}
