package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash is empty")
	}

	if hash == password {
		t.Fatal("Hash is same as password")
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("Password check failed")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Fatal("Password check succeeded with wrong password")
	}
}

func TestJWT(t *testing.T) {
	userID := "user-123"
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Fatal("Token is empty")
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	_, err = ValidateJWT("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}
