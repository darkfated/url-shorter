package memory

import (
	"context"
	"sync"

	"url-shorter/internal/domain"
	"url-shorter/internal/service"
)

type Store struct {
	mu       sync.RWMutex
	byShort  map[string]domain.Link
	byOrigin map[string]domain.Link
}

func New() *Store {
	return &Store{
		byShort:  make(map[string]domain.Link),
		byOrigin: make(map[string]domain.Link),
	}
}

func (s *Store) CreateLink(_ context.Context, link domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byOrigin[link.OriginalURL]; ok {
		return service.ErrDuplicateOriginalURL
	}
	if _, ok := s.byShort[link.ShortCode]; ok {
		return service.ErrDuplicateShortCode
	}

	s.byShort[link.ShortCode] = link
	s.byOrigin[link.OriginalURL] = link
	return nil
}

func (s *Store) FindByOriginalURL(_ context.Context, originalURL string) (domain.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.byOrigin[originalURL]
	if !ok {
		return domain.Link{}, service.ErrNotFound
	}
	return link, nil
}

func (s *Store) FindByShortCode(_ context.Context, shortCode string) (domain.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.byShort[shortCode]
	if !ok {
		return domain.Link{}, service.ErrNotFound
	}
	return link, nil
}
