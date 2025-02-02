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

func ValidOTP(otp string) bool {
	otp = strings.TrimSpace(otp)
	pattern := `^[A-Za-z0-9]{6}$` // Matches exactly 6 alphanumeric characters
	re := regexp.MustCompile(pattern)
	return re.MatchString(otp)
}
