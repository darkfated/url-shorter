package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/darkfated/url-shorter/internal/domain"
	"github.com/darkfated/url-shorter/internal/service"
)

func TestCreateAndFind(t *testing.T) {
	store := New()
	link := domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	}

	if err := store.CreateLink(context.Background(), link); err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	got, err := store.FindByOriginalURL(context.Background(), link.OriginalURL)
	if err != nil {
		t.Fatalf("FindByOriginalURL returned error: %v", err)
	}
	if got.ShortCode != link.ShortCode {
		t.Fatalf("unexpected short code %q", got.ShortCode)
	}

	got, err = store.FindByShortCode(context.Background(), link.ShortCode)
	if err != nil {
		t.Fatalf("FindByShortCode returned error: %v", err)
	}
	if got.OriginalURL != link.OriginalURL {
		t.Fatalf("unexpected original url %q", got.OriginalURL)
	}
}

func TestRejectDuplicates(t *testing.T) {
	store := New()
	link := domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	}

	if err := store.CreateLink(context.Background(), link); err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	if err := store.CreateLink(context.Background(), link); !errors.Is(err, service.ErrDuplicateOriginalURL) {
		t.Fatalf("expected duplicate original error, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	store := New()

	_, err := store.FindByOriginalURL(context.Background(), "missing")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
