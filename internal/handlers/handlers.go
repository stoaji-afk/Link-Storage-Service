package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"service/internal/service"
)

type Handlers struct {
	linkService *service.LinkService
}

func NewHandlers(linkService *service.LinkService) *Handlers {
	return &Handlers{linkService: linkService}
}

// CreateLinkHandler обрабатывает создание короткой ссылки (POST /links)
func (h *Handlers) CreateLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		http.Error(w, "URL must start with http:// or https://", http.StatusBadRequest)
		return
	}

	shortCode, err := h.linkService.CreateShortLink(r.Context(), req.URL)
	if err != nil {
		http.Error(w, "Failed to create short link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"short_code": shortCode})
}

// GetOriginalURLHandler обрабатывает получение оригинальной ссылки (GET /links/{short_code})
func (h *Handlers) GetOriginalURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	shortCode := pathParts[2]

	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	link, err := h.linkService.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
	} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url":    link.OriginalURL,
		"visits": link.Visits,
	})
}

// ListLinksHandler обрабатывает получение списка ссылок с пагинацией (GET /links)
func (h *Handlers) ListLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	links, err := h.linkService.ListLinks(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to list links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

// DeleteLinkHandler обрабатывает удаление ссылки (DELETE /links/{short_code})
func (h *Handlers) DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	shortCode := pathParts[2]

	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	err := h.linkService.DeleteLink(r.Context(), shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
	} else {
			http.Error(w, "Failed to delete link: "+err.Error(), http.StatusInternalServerError)
	}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetLinkStatsHandler обрабатывает получение статистики по ссылке (GET /links/{short_code}/stats)
func (h *Handlers) GetLinkStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	shortCodePath := strings.TrimSuffix(path, "/stats")
	pathParts := strings.Split(shortCodePath, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	shortCode := pathParts[2]

	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	link, err := h.linkService.GetLinkStats(r.Context(), shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
	} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link)
}
