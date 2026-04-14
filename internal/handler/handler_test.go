package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"url-shorter/internal/service"
	"url-shorter/internal/storage/memory"
)

func TestCreateAndRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.New()
	svc := service.NewWithGenerator(store, func() string { return "code000001" })
	h := New(svc, "https://urls.yandex.ru") // в качестве примера

	server := httptest.NewServer(h.Routes())
	t.Cleanup(server.Close)

	body := []byte(`{"url":"https://yandex.ru"}`)
	resp, err := http.Post(server.URL+"/api/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var created struct {
		OriginalURL string `json:"original_url"`
		ShortURL    string `json:"short_url"`
		ShortCode   string `json:"short_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if created.ShortCode != "code000001" {
		t.Fatalf("unexpected short code %q", created.ShortCode)
	}
	if created.ShortURL != "https://urls.yandex.ru/code000001" {
		t.Fatalf("unexpected short url %q", created.ShortURL)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	getResp, err := client.Get(server.URL + "/" + created.ShortCode)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusFound {
		t.Fatalf("unexpected GET status: %d", getResp.StatusCode)
	}

	if location := getResp.Header.Get("Location"); location != "https://yandex.ru" {
		t.Fatalf("unexpected redirect location %q", location)
	}
}

func TestCreateShortLinkInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.New()
	svc := service.New(store)
	h := New(svc, "https://urls.yandex.ru")

	server := httptest.NewServer(h.Routes())
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/api/shorten", "application/json", strings.NewReader("{"))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "неверный json" {
		t.Fatalf("unexpected error message %q", body.Error)
	}
}
