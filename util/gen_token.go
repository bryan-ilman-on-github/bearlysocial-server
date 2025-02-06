package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Creates a secure random token using crypto/rand.
func GenerateToken(uid string) (string, error) {
	b := make([]byte, 32)

	// Fill the byte slice with cryptographically random bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Create a new SHA-256 hash instance with random bytes written into the hash function.
	hasher := sha256.New()
	hasher.Write(b)

	// Compute the hash and return its hexadecimal representation.
	hashpass := hex.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("%s::%s", uid, hashpass), nil // A token is made of a uid (email) and a 'hashpass'.
}
