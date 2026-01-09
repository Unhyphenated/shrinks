package auth

import (
    "errors"
    "os"
    "time"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID uint64
    Email  string
    jwt.RegisteredClaims  // Built-in struct with exp, iat, iss, etc.
}

func GenerateToken(userID uint64, email string) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return "", errors.New("JWT_SECRET is not set")
    }

    claims := &Claims{
        UserID: userID,
        Email: email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            Issuer: "shrinks",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(token string) (*Claims, error) {

}