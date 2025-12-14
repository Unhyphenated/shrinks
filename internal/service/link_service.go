package service

import (
	"context"
	"fmt"
	
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

type LinkService struct {
	Store storage.Store // The Store interface is the dependency
}

func NewLinkService(s storage.Store) *LinkService {
	return &LinkService{Store: s}
}

func (ls *LinkService) Shorten(ctx context.Context, longURL string) (string, error) {
	// Rate Limiting
	// Input Sanitation

	shortURL, err := ls.Store.SaveLink(ctx, longURL)
	if err != nil {
		return "", fmt.Errorf("failed to save link: %w", err)
	}
	return shortURL, nil
}

func (ls *LinkService) Redirect(ctx context.Context, shortURL string) (string, error) {
	link, err := ls.Store.GetLinkByCode(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get link by code: %w", err)
	}

	if link == nil {
		return "", fmt.Errorf("link not found")
	}

	// TODO: Implement s.Store.UpdateClickCount(ctx, link.ID)

    return link.LongURL, nil
}