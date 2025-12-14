package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
)

type Store interface {
	SaveLink(link model.Link) (uint64, error)
	GetLinkByCode(code string) (*model.Link, error)
	// We'll add a Close method to satisfy deferral in main
    Close()
}

type PostgresStore struct {
	Pool *pgxpool.Pool // We use the Pool directly from pgxpool
}

// NewPostgresStore initializes the Postgres database connection pool.
func NewPostgresStore(dbURL string) (*PostgresStore, error) {
    ctx := context.Background() 
    
    // pgxpool.New uses the URL and sets up the pool based on defaults (or config)
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    // Ping the database using the pool's health check
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    fmt.Println("Successfully initialized Postgres Connection Pool!")
	return &PostgresStore{Pool: pool}, nil
}

func (s *PostgresStore) Close() {
    s.Pool.Close() // The professional way to shut down a pool
}