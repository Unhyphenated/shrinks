package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID int
    Email  string
    jwt.RegisteredClaims  // Built-in struct with exp, iat, iss, etc.
}

