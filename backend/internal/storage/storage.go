package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUniqueViolation = errors.New("unique violation")
	ErrLinkNotFound    = errors.New("link not found")
	ErrNotOwner        = errors.New("not owner")
)

type Closer interface {
	Close()
}

type LinkStore interface {
	Closer
	SaveLink(ctx context.Context, longURL string, userID *uint64) (string, error)
	GetLinkByCode(ctx context.Context, code string) (*model.Link, error)
	GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error)
	DeleteLink(ctx context.Context, shortCode string, userID uint64) error
	GetTotalLinks(ctx context.Context) (int, error)
	GetTotalRequests(ctx context.Context) (int, error)
}

type AuthStore interface {
	Closer
	CreateUser(ctx context.Context, email string, passwordHash string) (uint64, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uint64) (*model.User, error)

	// Refresh token methods
	CreateRefreshToken(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteUserRefreshTokens(ctx context.Context, userID uint64) error
}

type AnalyticsStore interface {
	Closer
	SaveAnalyticsEvent(ctx context.Context, event *model.AnalyticsEvent) error
	GetAnalyticsEvents(ctx context.Context, linkID uint64, period time.Time) ([]*model.AnalyticsEvent, error)
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

func (s *PostgresStore) SaveLink(ctx context.Context, longURL string, userID *uint64) (string, error) {
	var shortCode string

	var nextID uint64
	err := s.Pool.QueryRow(ctx, "SELECT nextval('links_id_seq')").Scan(&nextID)
	if err != nil {
		return "", err
	}

	// Start IDs at 100000 for better looking short codes
	if nextID < 100000 {
		nextID = 100000
		// Set the sequence to this value
		_, err = s.Pool.Exec(ctx, "SELECT setval('links_id_seq', $1, false)", nextID)
		if err != nil {
			return "", fmt.Errorf("failed to set sequence value: %w", err)
		}
	}

	// 2. Generate the code in Go
	shortCode = encoding.Encode(nextID)

	// 3. Insert the full record
	// Note: userID (as *uint64) will be NULL in DB if the pointer is nil
	insertQuery := `
        INSERT INTO links (id, long_url, short_code, user_id) 
        VALUES ($1, $2, $3, $4);
    `
	_, err = s.Pool.Exec(ctx, insertQuery, nextID, longURL, shortCode, userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert link: %w", err)
	}

	return shortCode, nil
}

func (s *PostgresStore) GetLinkByCode(ctx context.Context, shortCode string) (*model.Link, error) {
	query := `
		SELECT id, user_id, long_url, short_code, created_at
		FROM links 
		WHERE short_code = $1;
	`
	link := &model.Link{}
	err := s.Pool.QueryRow(ctx, query, shortCode).Scan(
		&link.ID,
		&link.UserID,
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

func (s *PostgresStore) GetUserByID(ctx context.Context, id uint64) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1;
	`

	user := &model.User{}
	err := s.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Return nil link and nil error for "not found"
		}
		return nil, fmt.Errorf("error querying user by id: %w", err)
	}

	return user, nil
}

func (s *PostgresStore) CreateRefreshToken(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := s.Pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	token := &model.RefreshToken{}

	err := s.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("error querying refresh token: %w", err)
	}
	return token, nil
}

func (s *PostgresStore) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := s.Pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (s *PostgresStore) DeleteUserRefreshTokens(ctx context.Context, userID uint64) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := s.Pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}

	return nil
}

func (s *PostgresStore) SaveAnalyticsEvent(ctx context.Context, event *model.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics (link_id, ip_address, user_agent, device_type, browser, os, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.Pool.Exec(ctx, query,
		event.LinkID,
		event.IPAddress,
		event.UserAgent,
		event.DeviceType,
		event.Browser,
		event.OS,
		event.ClickedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save analytics event: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetAnalyticsEvents(ctx context.Context, linkID uint64, period time.Time) ([]*model.AnalyticsEvent, error) {
	query := `
		SELECT id, link_id, ip_address::text, user_agent, device_type, browser, os, clicked_at
		FROM analytics
		WHERE link_id = $1 AND clicked_at > $2
	`
	rows, err := s.Pool.Query(ctx, query, linkID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics events: %w", err)
	}

	defer rows.Close()

	events := []*model.AnalyticsEvent{}
	for rows.Next() {
		var event model.AnalyticsEvent
		err := rows.Scan(
			&event.ID,
			&event.LinkID,
			&event.IPAddress,
			&event.UserAgent,
			&event.DeviceType,
			&event.Browser,
			&event.OS,
			&event.ClickedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics event: %w", err)
		}
		events = append(events, &event)
	}
	return events, nil
}

func (s *PostgresStore) DeleteLink(ctx context.Context, shortCode string, userID uint64) error {
	link, err := s.GetLinkByCode(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to get link by code: %w", err)
	}
	if link == nil {
		return ErrLinkNotFound
	}
	if link.UserID != nil && *link.UserID != userID {
		return ErrNotOwner
	}

	tx, err := s.Pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	query := `
		DELETE FROM analytics
		WHERE link_id = $1
	`
	_, err = tx.Exec(ctx, query, link.ID)
	if err != nil {
		return fmt.Errorf("failed to delete analytics events: %w", err)
	}

	query = `
		DELETE FROM links
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, query, link.ID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error) {
	query := `
		WITH total AS (
		SELECT count(*) as amount FROM links WHERE user_id = $1
		)
		SELECT id, user_id, long_url, short_code, created_at, total.amount
		FROM links, total
		WHERE user_id = $1
		ORDER BY created_at desc
		LIMIT $2 OFFSET $3
	`

	rows, err := s.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user links: %w", err)
	}

	defer rows.Close()

	links := []model.Link{}
	total := 0

	for rows.Next() {
		var link model.Link
		err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.LongURL,
			&link.ShortCode,
			&link.CreatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}

	return links, total, nil
}

func (s *PostgresStore) GetTotalLinks(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM links`
	var total int
	err := s.Pool.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total links: %w", err)
	}
	return total, nil
}

func (s *PostgresStore) GetTotalRequests(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM analytics`
	var total int
	err := s.Pool.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total requests: %w", err)
	}
	return total, nil
}
