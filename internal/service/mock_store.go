package service

import (
	"context"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)


type MockStore struct {
	SaveLinkFn    func(ctx context.Context, longURL string) (string, error)
	GetLinkByCodeFn func(ctx context.Context, shortURL string) (*model.Link, error)
	UpdateClickCountFn func(ctx context.Context, linkID uint64) error
	CloseFn       func()
}

var _ storage.Store = (*MockStore)(nil)

func (m *MockStore) SaveLink(ctx context.Context, longURL string) (string, error) {
	return m.SaveLinkFn(ctx, longURL) 
}

func (m *MockStore) GetLinkByCode(ctx context.Context, shortURL string) (*model.Link, error) {
	return m.GetLinkByCodeFn(ctx, shortURL)
}

func (m *MockStore) UpdateClickCount(ctx context.Context, linkID uint64) error {
	return m.UpdateClickCountFn(ctx, linkID)
}

func (m *MockStore) Close() {
	m.CloseFn()
}