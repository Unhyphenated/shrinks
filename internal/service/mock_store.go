package service

import (
	"context"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)


type MockStore struct {
	SaveLinkFn    func(ctx context.Context, longURL string) (string, error)
	GetLinkByCodeFn func(ctx context.Context, shortURL string) (*model.Link, error)
	CloseFn       func()
}

var _ storage.LinkStore = (*MockStore)(nil)

func (m *MockStore) SaveLink(ctx context.Context, longURL string) (string, error) {
	return m.SaveLinkFn(ctx, longURL) 
}

func (m *MockStore) GetLinkByCode(ctx context.Context, shortURL string) (*model.Link, error) {
	return m.GetLinkByCodeFn(ctx, shortURL)
}

func (m *MockStore) Close() {
	m.CloseFn()
}