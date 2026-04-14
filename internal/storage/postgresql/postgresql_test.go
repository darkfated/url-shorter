package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	"github.com/darkfated/url-shorter/internal/domain"
	"github.com/darkfated/url-shorter/internal/service"
)

func TestFindByOriginalURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	rows := sqlmock.NewRows([]string{"original_url", "short_code"}).
		AddRow("https://yandex.ru", "code000001")
	mock.ExpectQuery(`SELECT original_url, short_code FROM links WHERE original_url = \$1`).
		WithArgs("https://yandex.ru").
		WillReturnRows(rows)

	store := &Store{db: db}
	link, err := store.FindByOriginalURL(context.Background(), "https://yandex.ru")
	if err != nil {
		t.Fatalf("FindByOriginalURL returned error: %v", err)
	}
	if link.ShortCode != "code000001" {
		t.Fatalf("unexpected short code %q", link.ShortCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateLinkDuplicateOriginal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(`INSERT INTO links \(original_url, short_code\) VALUES \(\$1, \$2\)`).
		WithArgs("https://yandex.ru", "code000001").
		WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectQuery(`SELECT original_url, short_code FROM links WHERE original_url = \$1`).
		WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"original_url", "short_code"}).
			AddRow("https://yandex.ru", "code000009"))

	store := &Store{db: db}
	err = store.CreateLink(context.Background(), domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	})
	if !errors.Is(err, service.ErrDuplicateOriginalURL) {
		t.Fatalf("expected duplicate original error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCreateLinkSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(`INSERT INTO links \(original_url, short_code\) VALUES \(\$1, \$2\)`).
		WithArgs("https://yandex.ru", "code000001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	store := &Store{db: db}
	if err := store.CreateLink(context.Background(), domain.Link{
		OriginalURL: "https://yandex.ru",
		ShortCode:   "code000001",
	}); err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNewEmptyDSN(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatalf("expected error for empty dsn")
	}
}
