package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

// --- API Response Helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- CORS Middleware ---

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// --- Prompt Handlers ---

func (app *App) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20
	search := ""

	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
			if limit > 100 {
				limit = 100
			}
		}
	}
	if s := r.URL.Query().Get("search"); s != "" {
		search = strings.TrimSpace(s)
	}

	offset := (page - 1) * limit
	entries, total, err := app.storage.ListAllPrompts(offset, limit, search)
	if err != nil {
		slog.Error("failed to list prompts", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list prompts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": entries,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (app *App) handleGetPrompt(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.PathValue("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chat_id")
		return
	}

	entry, err := app.storage.GetPrompts(chatID)
	if err != nil {
		slog.Error("failed to get prompt", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get prompt")
		return
	}
	if entry == nil {
		writeError(w, http.StatusNotFound, "prompt not found")
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (app *App) handleUpdatePrompt(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.PathValue("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chat_id")
		return
	}

	var req struct {
		ImagePrompt string `json:"imagePrompt"`
		VideoPrompt string `json:"videoPrompt"`
		Provider    string `json:"provider"`
		ModelName   string `json:"modelName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := app.storage.UpdatePrompts(chatID, req.ImagePrompt, req.VideoPrompt, req.Provider, req.ModelName); err != nil {
		slog.Error("failed to update prompt", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update prompt")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"chatId":      chatID,
		"imagePrompt": req.ImagePrompt,
		"videoPrompt": req.VideoPrompt,
		"provider":    req.Provider,
		"modelName":   req.ModelName,
	})
}

func (app *App) handleDeletePrompt(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.PathValue("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chat_id")
		return
	}

	if err := app.storage.DeletePrompts(chatID); err != nil {
		slog.Error("failed to delete prompt", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete prompt")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (app *App) handleGetStats(w http.ResponseWriter, r *http.Request) {
	totalChats, imageCount, videoCount, totalPrompts, err := app.storage.GetStats()
	if err != nil {
		slog.Error("failed to get stats", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalChats":   totalChats,
		"imageCount":   imageCount,
		"videoCount":   videoCount,
		"totalPrompts": totalPrompts,
	})
}

func (app *App) handleListProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"providers": app.providers.GetProviderInfo(),
		"default":   app.providers.GetDefault(),
	})
}

// --- Template Handlers ---

func (app *App) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	entries, err := app.storage.ListTemplates()
	if err != nil {
		slog.Error("failed to list templates", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list templates")
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

func (app *App) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ImagePrompt string `json:"imagePrompt"`
		VideoPrompt string `json:"videoPrompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	id, err := app.storage.CreateTemplate(req.Name, req.Description, req.ImagePrompt, req.VideoPrompt)
	if err != nil {
		slog.Error("failed to create template", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create template")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id": id,
	})
}

func (app *App) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ImagePrompt string `json:"imagePrompt"`
		VideoPrompt string `json:"videoPrompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := app.storage.UpdateTemplate(id, req.Name, req.Description, req.ImagePrompt, req.VideoPrompt); err != nil {
		slog.Error("failed to update template", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update template")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (app *App) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := app.storage.DeleteTemplate(id); err != nil {
		slog.Error("failed to delete template", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete template")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// --- SPA Handler ---

func (app *App) handleSPA(w http.ResponseWriter, r *http.Request) {
	// API routes should not reach here.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "api endpoint not found")
		return
	}

	// Serve static files from embedded dist.
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Check if file exists in embedded FS.
	data, err := webFS.ReadFile("dashboard/dist/" + path)
	if err != nil {
		// Fallback to index.html for SPA routes.
		data, err = webFS.ReadFile("dashboard/dist/index.html")
		if err != nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}

	// Set content type based on extension.
	contentType := "application/octet-stream"
	switch {
	case strings.HasSuffix(path, ".html"):
		contentType = "text/html"
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".json"):
		contentType = "application/json"
	case strings.HasSuffix(path, ".png"):
		contentType = "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		contentType = "image/jpeg"
	case strings.HasSuffix(path, ".svg"):
		contentType = "image/svg+xml"
	case strings.HasSuffix(path, ".woff2"):
		contentType = "font/woff2"
	case strings.HasSuffix(path, ".woff"):
		contentType = "font/woff"
	case strings.HasSuffix(path, ".ttf"):
		contentType = "font/ttf"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// --- API Router Setup ---

func (app *App) setupAPIRoutes(mux *http.ServeMux) {
	// Prompts
	mux.HandleFunc("GET /api/prompts", withCORS(app.handleListPrompts))
	mux.HandleFunc("GET /api/prompts/{chatId}", withCORS(app.handleGetPrompt))
	mux.HandleFunc("PUT /api/prompts/{chatId}", withCORS(app.handleUpdatePrompt))
	mux.HandleFunc("DELETE /api/prompts/{chatId}", withCORS(app.handleDeletePrompt))
	mux.HandleFunc("GET /api/stats", withCORS(app.handleGetStats))

	// Providers
	mux.HandleFunc("GET /api/providers", withCORS(app.handleListProviders))

	// Templates
	mux.HandleFunc("GET /api/templates", withCORS(app.handleListTemplates))
	mux.HandleFunc("POST /api/templates", withCORS(app.handleCreateTemplate))
	mux.HandleFunc("PUT /api/templates/{id}", withCORS(app.handleUpdateTemplate))
	mux.HandleFunc("DELETE /api/templates/{id}", withCORS(app.handleDeleteTemplate))

	// Health check
	mux.HandleFunc("GET /health", withCORS(app.handleHealth))

	// SPA catch-all (must be last)
	mux.HandleFunc("/", app.handleSPA)
}
