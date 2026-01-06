package auth

import (
	"context"
	"errors"
	"net/mail"

	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidEmail    = errors.New("invalid email format")
    ErrPasswordTooShort = errors.New("password must be at least 8 characters")
    ErrPasswordTooLong  = errors.New("password exceeds 72 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService struct {
	Store storage.AuthStore
}

func NewAuthService(s storage.AuthStore) *AuthService {
	return &AuthService{Store: s}
}

func (as *AuthService) Register(ctx context.Context, email string, password string) (uint64, error) {
	if (len(password) < 8) {
		return 0, ErrPasswordTooShort
	}

	if (len(password) > 72) {
		return 0, ErrPasswordTooLong
	}

	_, err := mail.ParseAddress(email) 
	if err != nil {
		return 0, ErrInvalidEmail
	}

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
	_, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	user, err := as.Store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}
	return "", nil
}