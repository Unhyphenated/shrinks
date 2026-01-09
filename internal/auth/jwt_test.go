package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}
}

func TestGenerateToken_MissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	_, err := GenerateToken(123, "test@example.com")
	if err == nil {
		t.Error("Expected error when JWT_SECRET is not set")
	}
}

func TestValidateToken_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	userID := uint64(123)
	email := "test@example.com"

	token, err := GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not.a.valid.token"},
		{"wrong format", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token")
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")

	token, err := GenerateToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Change secret
	os.Setenv("JWT_SECRET", "wrong-secret")

	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret")
	}

	os.Unsetenv("JWT_SECRET")
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	// Create an expired token manually
	secret := []byte("test-secret-123")
	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "shrinks",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected error for expired token")
	}
}

func TestValidateToken_MissingSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	token, _ := GenerateToken(123, "test@example.com")

	os.Unsetenv("JWT_SECRET")

	_, err := ValidateToken(token)
	if err == nil {
		t.Error("Expected error when JWT_SECRET is not set")
	}
}

func TestJWT_RoundTrip(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	// Generate multiple tokens and verify they're unique
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		token, err := GenerateToken(uint64(i), "test@example.com")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if tokens[token] {
			t.Errorf("Duplicate token generated: %s", token)
		}
		tokens[token] = true

		// Verify each token is valid
		claims, err := ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}
		if claims.UserID != uint64(i) {
			t.Errorf("Expected UserID %d, got %d", i, claims.UserID)
		}
	}
}
