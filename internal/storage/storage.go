package storage

import (
	"context"
	"fmt"

    "github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	SaveLink(ctx context.Context, longURL string) (string, error)
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

func (s *PostgresStore) SaveLink(ctx context.Context, longURL string) (string, error) {
	// 1. Begin the Transaction using the pool
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) 

	// 2. Insert URL and get ID
	var generatedID uint64
	insertQuery := `
		INSERT INTO links (long_url, short_url) 
		VALUES ($1, '')
		RETURNING id;
	`
	err = tx.QueryRow(ctx, insertQuery, longURL).Scan(&generatedID)
	if err != nil {
		return "", fmt.Errorf("transaction insert failed: %w", err)
	}

	// 3. Encode the ID
	shortURL := encoding.Encode(generatedID) 

	// 4. Update the row with the short code
	updateQuery := `
		UPDATE links 
		SET short_code = $1 
		WHERE id = $2;
	`
	_, err = tx.Exec(ctx, updateQuery, shortURL, generatedID)
	if err != nil {
		return "", fmt.Errorf("transaction update failed: %w", err)
	}

	// 5. Commit the Transaction
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("transaction commit failed: %w", err)
	}

	return shortURL, nil
}