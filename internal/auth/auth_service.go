package auth

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidEmail    = errors.New("invalid email format")
    ErrPasswordTooShort = errors.New("password must be at least 8 characters")
    ErrPasswordTooLong  = errors.New("password exceeds 72 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
    ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrDeletingRefreshToken = errors.New("failed to delete refresh token")
)

type AuthProvider interface {
	Register(ctx context.Context, email string, password string) (model.RegisterResponse, error)
	Login(ctx context.Context, email string, password string) (model.AuthResponse, error)
    RefreshAccessToken(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error)
    Logout(ctx context.Context, refreshToken string) error
}

type AuthService struct {
	Store storage.AuthStore
}

func NewAuthService(s storage.AuthStore) *AuthService {
	return &AuthService{Store: s}
}

func (as *AuthService) Register(ctx context.Context, email string, password string) (model.RegisterResponse, error) {
	if len(password) < 8 {
		return model.RegisterResponse{}, ErrPasswordTooShort
	}

	if len(password) > 72 {
		return model.RegisterResponse{}, ErrPasswordTooLong
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return model.RegisterResponse{}, ErrInvalidEmail
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return model.RegisterResponse{}, err
	}

	userID, err := as.Store.CreateUser(ctx, email, string(passwordHash))

	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			return model.RegisterResponse{}, ErrUserAlreadyExists
		}
		return model.RegisterResponse{}, err
	}

	
	return model.RegisterResponse{
		UserID: userID,
	}, nil
}

func (as *AuthService) Login(ctx context.Context, email string, password string) (model.AuthResponse, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return model.AuthResponse{}, ErrInvalidEmail
	}

	user, err := as.Store.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	accessToken, err := GenerateToken(user.ID, email)
	if err != nil {
		return model.AuthResponse{}, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return model.AuthResponse{}, err
	}

	tokenHash := HashRefreshToken(refreshToken)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err = as.Store.CreateRefreshToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		User: *user,
	}, nil
}

func (as *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error) {
	tokenHash := HashRefreshToken(refreshToken)

	storedToken, err := as.Store.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	if storedToken == nil {
		return model.RefreshTokenResponse{}, ErrInvalidRefreshToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return model.RefreshTokenResponse{}, ErrRefreshTokenExpired
	}

	user, err := as.Store.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	if user == nil {
		return model.RefreshTokenResponse{}, ErrInvalidRefreshToken
	}

	accessToken, err := GenerateToken(user.ID, user.Email)
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	return model.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}	

func (as *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := HashRefreshToken(refreshToken)
	token, err := as.Store.GetRefreshToken(ctx, tokenHash)
	
	if err != nil {
		return errors.New("failure to get refresh token")
	}

	if token == nil {
		return ErrRefreshTokenNotFound
	}

	if err := as.Store.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return ErrDeletingRefreshToken
	}

	return nil
}