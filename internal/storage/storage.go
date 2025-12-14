package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	SaveLink(ctx context.Context, LongURL string) (string, error)
    GetLinkByCode(ctx context.Context, code string) (*model.Link, error)
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
    s.Pool.Close()
}

func (s *PostgresStore) SaveLink(longURL string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tx, err := s.Pool.Begin(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer tx.Rollback(ctx)
    
    var generatedID uint64
    insertQuery := `
		INSERT INTO links (long_url, short_url) 
		VALUES ($1, '') -- Insert with a blank short_code first
		RETURNING id;
	`

    err = tx.QueryRow(ctx, insertQuery, longURL).Scan(&generatedID)
    if err != nil {
        return "", fmt.Errorf("failed to insert link: %w", err)
    }

    shortURL := encoding.Encode(generatedID)
    
    updateQuery := `
        UPDATE links 
        SET short_url = $1 
        WHERE id = $2;
    `
    
    _, err = tx.Exec(ctx, updateQuery, shortURL, generatedID)
    if err != nil {
        return "", fmt.Errorf("failed to update short_url: %w", err)
    }

    err = tx.Commit(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to commit transaction: %w", err)
    }

    return shortURL, nil
}