package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUniqueViolation = errors.New("unique violation")
)

type LinkStore interface {
	SaveLink(ctx context.Context, longURL string) (string, error)
    GetLinkByCode(ctx context.Context, code string) (*model.Link, error)
    Close()
}

type AuthStore interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (uint64, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
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
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) 

	var generatedID uint64
	insertQuery := `
		INSERT INTO links (long_url, short_code) 
		VALUES ($1, '')
		RETURNING id;
	`
	err = tx.QueryRow(ctx, insertQuery, longURL).Scan(&generatedID)
	if err != nil {
		return "", fmt.Errorf("transaction insert failed: %w", err)
	}

	shortURL := encoding.Encode(generatedID) 

	updateQuery := `
		UPDATE links 
		SET short_code = $1 
		WHERE id = $2;
	`
	_, err = tx.Exec(ctx, updateQuery, shortURL, generatedID)
	if err != nil {
		return "", fmt.Errorf("transaction update failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("transaction commit failed: %w", err)
	}

	return shortURL, nil
}

func (s *PostgresStore) GetLinkByCode(ctx context.Context, shortURL string) (*model.Link, error) {
	query := `
		SELECT id, long_url, short_code, created_at
		FROM links 
		WHERE short_code = $1;
	`
	link := &model.Link{}
	err := s.Pool.QueryRow(ctx, query, shortURL).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.CreatedAt,
	)

	if err != nil {
        // Handle the specific, common case where the code isn't in the database.
        if err == pgx.ErrNoRows {
            return nil, nil // Return nil link and nil error for "not found"
        }
	        return nil, fmt.Errorf("error querying link by code: %w", err)
    }

    return link, nil
}

func (s *PostgresStore) CreateUser(ctx context.Context, email string, passwordHash string) (uint64, error) {
	var generatedID uint64
	insertQuery := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING
		RETURNING id;
	`

	err := s.Pool.QueryRow(ctx, insertQuery, email, passwordHash).Scan(&generatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			existingUser, lookupErr := s.GetUserByEmail(ctx, email)
			if lookupErr != nil {
				return 0, fmt.Errorf("error looking up user: %w", lookupErr)
			}
			if existingUser != nil {
				return 0, ErrUniqueViolation
			}
			return 0, fmt.Errorf("error inserting user: %w", err)
		}
		return 0, fmt.Errorf("transaction insert failed: %w", err)
	}

	return generatedID, nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1;
	`

	user := &model.User{}
	err := s.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
            return nil, nil // Return nil link and nil error for "not found"
        }
		return nil, fmt.Errorf("error querying user by email: %w", err)
	}

	return user, nil
}