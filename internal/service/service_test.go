package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/darkfated/url-shorter/internal/domain"
)

type stubStore struct {
	byOriginal map[string]domain.Link
	byShort    map[string]domain.Link
	createFn   func(link domain.Link) error
}

func newStubStore() *stubStore {
	return &stubStore{
		byOriginal: make(map[string]domain.Link),
		byShort:    make(map[string]domain.Link),
	}
}

func (s *stubStore) CreateLink(_ context.Context, link domain.Link) error {
	if s.createFn != nil {
		if err := s.createFn(link); err != nil {
			return err
		}
	}
	s.byOriginal[link.OriginalURL] = link
	s.byShort[link.ShortCode] = link
	return nil
}

func (s *stubStore) FindByOriginalURL(_ context.Context, originalURL string) (domain.Link, error) {
	link, ok := s.byOriginal[originalURL]
	if !ok {
		return domain.Link{}, ErrNotFound
	}
	return link, nil
}

func (s *stubStore) FindByShortCode(_ context.Context, shortCode string) (domain.Link, error) {
	link, ok := s.byShort[shortCode]
	if !ok {
		return domain.Link{}, ErrNotFound
	}
	return link, nil
}

func TestCreateExistingLink(t *testing.T) {
	store := newStubStore()
	store.byOriginal["https://yandex.ru"] = domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	}

	svc := NewWithGenerator(store, func() string { return "code000002" })

	link, err := svc.Create(context.Background(), "https://yandex.ru")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if link.ShortCode != "code000001" {
		t.Fatalf("expected existing short code, got %q", link.ShortCode)
	}
}

func TestCreateLink(t *testing.T) {
	store := newStubStore()
	codes := []string{"code000001"}
	svc := NewWithGenerator(store, func() string {
		return codes[0]
	})

	link, err := svc.Create(context.Background(), "https://yandex.ru")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if link.ShortCode != "code000001" {
		t.Fatalf("unexpected short code %q", link.ShortCode)
	}
	if _, ok := store.byOriginal["https://yandex.ru"]; !ok {
		t.Fatalf("link was not saved")
	}
}

func TestRetryShortCode(t *testing.T) {
	store := newStubStore()
	store.createFn = func(link domain.Link) error {
		if link.ShortCode == "code000001" {
			return ErrDuplicateShortCode
		}
		return nil
	}

	codes := []string{"code000001", "code000002"}
	next := 0
	svc := NewWithGenerator(store, func() string {
		code := codes[next]
		next++
		return code
	})

	link, err := svc.Create(context.Background(), "https://yandex.ru")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if link.ShortCode != "code000002" {
		t.Fatalf("expected retry code, got %q", link.ShortCode)
	}
}

func TestResolveLink(t *testing.T) {
	store := newStubStore()
	store.byShort["code000001"] = domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	}

	svc := New(store)

	link, err := svc.Resolve(context.Background(), "code000001")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if link.OriginalURL != "https://yandex.ru" {
		t.Fatalf("unexpected original url %q", link.OriginalURL)
	}
}

func TestRejectInvalidURL(t *testing.T) {
	svc := New(newStubStore())

	_, err := svc.Create(context.Background(), "not-a-url")
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestRejectTooLongURL(t *testing.T) {
	svc := New(newStubStore())

	longURL := "https://yandex.ru/" + strings.Repeat("a", 300)
	_, err := svc.Create(context.Background(), longURL)
	if !errors.Is(err, ErrURLTooLong) {
		t.Fatalf("expected ErrURLTooLong, got %v", err)
	}
}
