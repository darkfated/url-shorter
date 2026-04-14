package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"strings"

	"url-shorter/internal/domain"
)

const shortCodeLength = 10
const shortCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
const maxOriginalURLLength = 260

var (
	ErrNotFound             = errors.New("ссылка не найдена")
	ErrDuplicateOriginalURL = errors.New("такая ссылка уже есть")
	ErrDuplicateShortCode   = errors.New("такая короткая ссылка уже есть")
	ErrInvalidURL           = errors.New("ссылка некорректная")
	ErrURLTooLong           = errors.New("ссылка слишком длинная")
)

type Store interface {
	CreateLink(ctx context.Context, link domain.Link) error
	FindByOriginalURL(ctx context.Context, originalURL string) (domain.Link, error)
	FindByShortCode(ctx context.Context, shortCode string) (domain.Link, error)
}

type Service struct {
	store   Store
	newCode func() string
}

func New(store Store) *Service {
	return &Service{
		store:   store,
		newCode: generateShortCode,
	}
}

func NewWithGenerator(store Store, newCode func() string) *Service {
	if newCode == nil {
		newCode = generateShortCode
	}
	return &Service{
		store:   store,
		newCode: newCode,
	}
}

func (s *Service) Create(ctx context.Context, originalURL string) (domain.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	if err := validateURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	existing, err := s.store.FindByOriginalURL(ctx, originalURL)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return domain.Link{}, err
	}

	for i := 0; i < 10; i++ {
		link := domain.Link{
			OriginalURL: originalURL,
			ShortCode:   s.newCode(),
		}

		err = s.store.CreateLink(ctx, link)
		if err == nil {
			return link, nil
		}
		if errors.Is(err, ErrDuplicateOriginalURL) {
			existing, findErr := s.store.FindByOriginalURL(ctx, originalURL)
			if findErr == nil {
				return existing, nil
			}
			if !errors.Is(findErr, ErrNotFound) {
				return domain.Link{}, findErr
			}
			continue
		}
		if errors.Is(err, ErrDuplicateShortCode) {
			continue
		}
		return domain.Link{}, err
	}

	return domain.Link{}, errors.New("не удалось создать короткую ссылку")
}

func (s *Service) Resolve(ctx context.Context, shortCode string) (domain.Link, error) {
	shortCode = strings.TrimSpace(shortCode)
	if shortCode == "" {
		return domain.Link{}, ErrNotFound
	}

	return s.store.FindByShortCode(ctx, shortCode)
}

func validateURL(raw string) error {
	if len(raw) > maxOriginalURLLength {
		return ErrURLTooLong
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return ErrInvalidURL
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

func generateShortCode() string {
	code := make([]byte, shortCodeLength)
	maxLen := big.NewInt(int64(len(shortCodeAlphabet)))

	for i := range code {
		n, err := rand.Int(rand.Reader, maxLen)
		if err != nil {
			panic(err)
		}
		code[i] = shortCodeAlphabet[n.Int64()]
	}

	return string(code)
}
