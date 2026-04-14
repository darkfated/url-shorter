package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"url-shorter/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.POST("/api/shorten", h.createShortLink)
	r.GET("/:code", h.resolveShortLink)

	return r
}

type createShortLinkRequest struct {
	URL string `json:"url"`
}

type createShortLinkResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ShortCode   string `json:"short_code"`
}

func (h *Handler) createShortLink(c *gin.Context) {
	var req createShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный json"})
		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.URL)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createShortLinkResponse{
		OriginalURL: link.OriginalURL,
		ShortURL:    shortURL(c, link.ShortCode),
		ShortCode:   link.ShortCode,
	})
}

func (h *Handler) resolveShortLink(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	link, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, gin.H{"error": "ссылка некорректная"})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
	}
}

func shortURL(c *gin.Context, code string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/" + code
}
