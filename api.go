package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gemini-telegram-bot/ai"
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
		TemplateID int64             `json:"templateId"`
		Provider   string            `json:"provider"`
		ModelName  string            `json:"modelName"`
		Params     map[string]any    `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing entry before update to detect changes
	oldEntry, _ := app.storage.GetPrompts(chatID)

	var paramsJSON string
	if req.Params != nil {
		b, err := json.Marshal(req.Params)
		if err == nil {
			paramsJSON = string(b)
		}
	}

	if err := app.storage.UpdatePrompts(chatID, req.TemplateID, req.Provider, req.ModelName, paramsJSON); err != nil {
		slog.Error("failed to update prompt", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update prompt")
		return
	}

	// Notify user on Telegram if provider or model changed
	if oldEntry == nil || oldEntry.Provider != req.Provider || oldEntry.ModelName != req.ModelName {
		var msg string
		if req.Provider != "" && req.ModelName != "" {
			msg = fmt.Sprintf("✅ AI model updated!\n🤖 Provider: `%s`\n📦 Model: `%s`", req.Provider, req.ModelName)
		} else if req.Provider != "" {
			msg = fmt.Sprintf("✅ AI provider updated to `%s`", req.Provider)
		}
		if msg != "" {
			if _, err := app.bot.SendMessage(chatID, msg); err != nil {
				slog.Error("failed to send model change notification", "chat_id", chatID, "error", err)
			}
		}
	}

	// Get updated entry to return resolved template name
	entry, err := app.storage.GetPrompts(chatID)
	if err != nil {
		slog.Error("failed to get updated prompt", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get updated prompt")
		return
	}

	writeJSON(w, http.StatusOK, entry)
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
	totalChats, templateCount, totalPrompts, err := app.storage.GetStats()
	if err != nil {
		slog.Error("failed to get stats", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalChats":    totalChats,
		"templateCount": templateCount,
		"totalPrompts":  totalPrompts,
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

// --- Generate Handler ---

func (app *App) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type             string         `json:"type"`
		Prompt           string         `json:"prompt"`
		Provider         string         `json:"provider"`
		ModelName        string         `json:"modelName"`
		Params           map[string]any `json:"params"`
		ReferenceImage   string         `json:"referenceImage"`
		ReferenceMimeType string        `json:"referenceMimeType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	if req.Type != "image" && req.Type != "video" {
		writeError(w, http.StatusBadRequest, "type must be 'image' or 'video'")
		return
	}

	providerName := req.Provider
	if providerName == "" {
		providerName = app.providers.GetDefault()
	}

	provider := app.providers.Get(providerName)
	if provider == nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("provider not found: %s", providerName))
		return
	}

	modelName := req.ModelName
	if modelName == "" {
		models := provider.GetModels()
		if len(models) > 0 {
			modelName = models[0]
		}
	}

	opts := ai.GenerateOptions{
		ModelName: modelName,
		Params:    req.Params,
	}

	var refData []byte
	var refMime string
	if req.ReferenceImage != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.ReferenceImage)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid reference image base64")
			return
		}
		refData = decoded
		refMime = req.ReferenceMimeType
		if refMime == "" {
			refMime = "image/jpeg"
		}
	}

	start := time.Now()
	ctx := context.Background()

	var data []byte
	var mimeType string
	var err error

	switch req.Type {
	case "image":
		data, err = provider.GenerateImage(ctx, req.Prompt, refData, refMime, opts)
		mimeType = "image/png"
	case "video":
		data, err = provider.GenerateVideo(ctx, req.Prompt, refData, refMime, opts)
		mimeType = "video/mp4"
	}

	durationMs := time.Since(start).Milliseconds()

	if err != nil {
		slog.Error("generation failed", "type", req.Type, "provider", providerName, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":      err.Error(),
			"durationMs": durationMs,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":         base64.StdEncoding.EncodeToString(data),
		"mimeType":     mimeType,
		"providerName": providerName,
		"modelName":    modelName,
		"durationMs":   durationMs,
	})
}

// --- SPA Handler ---

func (app *App) handleSPA(w http.ResponseWriter, r *http.Request) {
	// API routes should not reach here.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "api endpoint not found")
		return
	}

	// In development mode, webFS is nil - dashboard is served separately by Vite
	if webFS == nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h1>Development Mode</h1><p>Dashboard is served separately at <a href=\"http://localhost:8080\">http://localhost:8080</a></p>"))
		return
	}

	// Serve static files from embedded dist.
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Check if file exists in embedded FS.
	data, err := fs.ReadFile(webFS, "dashboard/dist/"+path)
	if err != nil {
		// Fallback to index.html for SPA routes.
		data, err = fs.ReadFile(webFS, "dashboard/dist/index.html")
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

// --- Generation History Handlers ---

func (app *App) handleListGenerationHistory(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20
	statusFilter := "all"
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
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = s
	}
	if s := r.URL.Query().Get("search"); s != "" {
		search = strings.TrimSpace(s)
	}

	offset := (page - 1) * limit
	entries, total, err := app.storage.ListGenerationHistory(offset, limit, statusFilter, search)
	if err != nil {
		slog.Error("failed to list generation history", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list generation history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": entries,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (app *App) handleClearGenerationHistory(w http.ResponseWriter, r *http.Request) {
	if err := app.storage.ClearGenerationHistory(); err != nil {
		slog.Error("failed to clear generation history", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to clear generation history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "cleared"})
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

	// Generation History
	mux.HandleFunc("GET /api/generations", withCORS(app.handleListGenerationHistory))
	mux.HandleFunc("DELETE /api/generations", withCORS(app.handleClearGenerationHistory))

	// Generation
	mux.HandleFunc("POST /api/generate", withCORS(app.handleGenerate))

	// Health check
	mux.HandleFunc("GET /health", withCORS(app.handleHealth))

	// SPA catch-all (must be last)
	mux.HandleFunc("/", app.handleSPA)
}
