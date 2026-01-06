package auth

import (
	"context"
	"errors"

	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidEmail    = errors.New("invalid email format")
    ErrPasswordTooShort = errors.New("password must be at least 8 characters")
    ErrPasswordTooLong  = errors.New("password exceeds 72 characters")
)

type AuthService struct {
	Store storage.AuthStore
}

func NewAuthService(s storage.AuthStore) *AuthService {
	return &AuthService{Store: s}
}

func (as *AuthService) Register(ctx context.Context, email string, password string) (uint64, error) {
	var generatedID uint64

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return 0, err
	}

	generatedID, err = as.Store.CreateUser(ctx, email, string (passwordHash))

	if err != nil {
		return 0, err
	}

	return generatedID, nil
}

func (as *AuthService) Login(ctx context.Context, email string, password string) (string, error) {

	return "", nil
}