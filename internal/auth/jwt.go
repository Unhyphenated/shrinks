package auth

import (
	"os"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID int
    Email  string
    jwt.RegisteredClaims  // Built-in struct with exp, iat, iss, etc.
}

func GenerateToken(userID int, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := &Claims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer: "shrinks",
			Subject: email,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
