package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword encrypts the given plaintext password using bcrypt
// The cost parameter controls the hash computation difficulty; the default value is sufficient
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword automatically generates a salt and includes it in the final hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword verifies if the plaintext password matches the bcrypt hash
// The hash itself contains the salt used during encryption, so no extra salt is needed for verification
func CheckPassword(password, hashedPassword string) error {
	// CompareHashAndPassword is a constant-time comparison function that prevents timing attacks
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
