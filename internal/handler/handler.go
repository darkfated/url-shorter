package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"url-shorter/internal/service"
)

type Handler struct {
	svc     *service.Service
	baseURL string
}

func New(svc *service.Service, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/")}
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

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) createShortLink(c *gin.Context) {
	var req createShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "неверный json")
		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.URL)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createShortLinkResponse{
		OriginalURL: link.OriginalURL,
		ShortURL:    shortURL(h.baseURL, link.ShortCode),
		ShortCode:   link.ShortCode,
	})
}

func (h *Handler) resolveShortLink(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	link, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func (h *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidURL):
		h.writeError(c, http.StatusBadRequest, "ссылка некорректная")
	case errors.Is(err, service.ErrURLTooLong):
		h.writeError(c, http.StatusBadRequest, "ссылка слишком длинная")
	case errors.Is(err, service.ErrNotFound):
		h.writeError(c, http.StatusNotFound, "не найдено")
	default:
		h.writeError(c, http.StatusInternalServerError, "внутренняя ошибка")
	}
}

func (h *Handler) writeError(c *gin.Context, status int, message string) {
	c.JSON(status, errorResponse{Error: message})
}

func shortURL(baseURL, code string) string {
	return baseURL + "/" + code
}
