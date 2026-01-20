package auth

import (
	"context"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
)

type MockAuthService struct {
	RegisterFn           func(ctx context.Context, email, password string) (model.RegisterResponse, error)
	LoginFn              func(ctx context.Context, email, password string) (model.AuthResponse, error)
	RefreshAccessTokenFn func(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error)
	LogoutFn             func(ctx context.Context, refreshToken string) error
}

func (m *MockAuthService) Register(ctx context.Context, email, password string) (model.RegisterResponse, error) {
	return m.RegisterFn(ctx, email, password)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (model.AuthResponse, error) {
	return m.LoginFn(ctx, email, password)
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error) {
	return m.RefreshAccessTokenFn(ctx, refreshToken)
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return m.LogoutFn(ctx, refreshToken)
}
