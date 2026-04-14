package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/darkfated/url-shorter/internal/domain"
	"github.com/darkfated/url-shorter/internal/service"
)

type Store struct {
	db *sql.DB
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS links (
	original_url TEXT NOT NULL UNIQUE,
	short_code CHAR(10) PRIMARY KEY
);
`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) CreateLink(ctx context.Context, link domain.Link) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO links (original_url, short_code) VALUES ($1, $2)`,
		link.OriginalURL, link.ShortCode,
	)
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		existing, findErr := s.FindByOriginalURL(ctx, link.OriginalURL)
		if findErr == nil && existing.ShortCode != "" {
			return service.ErrDuplicateOriginalURL
		}
		if errors.Is(findErr, service.ErrNotFound) {
			return service.ErrDuplicateShortCode
		}
		if findErr != nil {
			return findErr
		}
		return service.ErrDuplicateOriginalURL
	}

	return err
}

func (s *Store) FindByOriginalURL(ctx context.Context, originalURL string) (domain.Link, error) {
	var link domain.Link
	err := s.db.QueryRowContext(ctx,
		`SELECT original_url, short_code FROM links WHERE original_url = $1`,
		originalURL,
	).Scan(&link.OriginalURL, &link.ShortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, service.ErrNotFound
		}
		return domain.Link{}, err
	}
	return link, nil
}

func (s *Store) FindByShortCode(ctx context.Context, shortCode string) (domain.Link, error) {
	var link domain.Link
	err := s.db.QueryRowContext(ctx,
		`SELECT original_url, short_code FROM links WHERE short_code = $1`,
		shortCode,
	).Scan(&link.OriginalURL, &link.ShortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, service.ErrNotFound
		}
		return domain.Link{}, err
	}
	return link, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
