package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gemini-telegram-bot/ai"
	"gemini-telegram-bot/ai/providers/gemini"
	"gemini-telegram-bot/ai/providers/kling"
	"gemini-telegram-bot/ai/providers/recraft"
	"gemini-telegram-bot/storage"
	"gemini-telegram-bot/worker"
)

// App holds all application dependencies.
type App struct {
	config    *Config
	bot       *BotClient
	providers *ai.Registry
	pool      *worker.Pool
	storage   *storage.PromptStore
	jobMap    map[int64]*ActiveJob
}

// ActiveJob tracks an in-progress job for a chat.
type ActiveJob struct {
	ChatID    int64
	MessageID int
	JobType   string
}

func main() {
	// Setup structured logging.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Load configuration.
	config := LoadConfig()

	// Initialize dependencies.
	bot := NewBotClient(config.BotToken)

	ctx := context.Background()

	// Initialize AI providers.
	providers := ai.NewRegistry()

	geminiClient, err := gemini.NewClient(ctx, config.GeminiAPIKey)
	if err != nil {
		slog.Error("failed to initialize gemini client", "error", err)
		os.Exit(1)
	}
	providers.Register(geminiClient)

	if config.KlingAPIKey != "" {
		klingClient := kling.NewClient(config.KlingAPIKey)
		providers.Register(klingClient)
	}

	if config.RecraftAPIKey != "" {
		recraftClient := recraft.NewClient(config.RecraftAPIKey)
		providers.Register(recraftClient)
	}

	store, err := storage.NewPromptStore(config.DatabasePath)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	app := &App{
		config:    config,
		bot:       bot,
		providers: providers,
		storage:   store,
		jobMap:    make(map[int64]*ActiveJob),
	}

	// Create worker pool.
	app.pool = worker.NewPool(config.WorkerPoolSize, app.handleJob)

	// Setup HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /telegram", app.handleWebhook)
	app.setupAPIRoutes(mux)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		app.pool.Shutdown()
	}()

	slog.Info("server starting", "port", config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world from go"))
}

func (app *App) handleWebhook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var update struct {
		Message struct {
			MessageID int64 `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			Text            string `json:"text"`
			Caption         string `json:"caption"`
			Photo           []struct {
				FileID string `json:"file_id"`
			} `json:"photo"`
			ReplyToMessage *struct {
				Photo []struct {
					FileID string `json:"file_id"`
				} `json:"photo"`
			} `json:"reply_to_message"`
		} `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		slog.Error("failed to decode webhook", "error", err)
		return
	}

	message := update.Message
	if message.MessageID == 0 {
		return
	}

	chatID := message.Chat.ID
	text := message.Text
	caption := message.Caption

	// Handle /set-template
	if strings.HasPrefix(text, "/set-template") {
		args := strings.TrimSpace(strings.TrimPrefix(text, "/set-template"))
		if args == "" {
			app.bot.SendMessage(chatID, "Usage: /set-template <template_id>\nAssigns a template to this chat for image/video generation.")
			return
		}
		templateID, err := strconv.ParseInt(args, 10, 64)
		if err != nil {
			app.bot.SendMessage(chatID, "❌ Invalid template ID. Use a number.")
			return
		}

		// Verify template exists.
		template, err := app.storage.GetTemplate(templateID)
		if err != nil {
			slog.Error("failed to get template", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to verify template.")
			return
		}
		if template == nil {
			app.bot.SendMessage(chatID, fmt.Sprintf("❌ Template not found: %d", templateID))
			return
		}

		if err := app.storage.SetTemplate(chatID, templateID); err != nil {
			slog.Error("failed to set template", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to save template assignment.")
			return
		}
		app.bot.SendMessage(chatID, fmt.Sprintf("✅ Template set to: %s", template.Name))
		return
	}

	// Handle /templates
	if strings.HasPrefix(text, "/templates") {
		templates, err := app.storage.ListTemplates()
		if err != nil {
			slog.Error("failed to list templates", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to load templates.")
			return
		}
		if len(templates) == 0 {
			app.bot.SendMessage(chatID, "No templates available. Create templates from the dashboard.")
			return
		}
		var b strings.Builder
		b.WriteString("📋 *Available Templates*\n\n")
		for _, t := range templates {
			hasImage := ""
			hasVideo := ""
			if t.ImagePrompt != "" {
				hasImage = " 🖼️"
			}
			if t.VideoPrompt != "" {
				hasVideo = " 🎬"
			}
			b.WriteString(fmt.Sprintf("*%d*. %s%s%s\n", t.ID, t.Name, hasImage, hasVideo))
			if t.Description != "" {
				b.WriteString(fmt.Sprintf("   _%s_\n", t.Description))
			}
		}
		b.WriteString("\nUse `/set-template <id>` to assign one to this chat.")
		app.bot.SendMessage(chatID, b.String())
		return
	}

	// Handle /set-provider
	if strings.HasPrefix(text, "/set-provider") {
		args := strings.TrimSpace(strings.TrimPrefix(text, "/set-provider"))
		parts := strings.Fields(args)
		if len(parts) < 2 {
			app.bot.SendMessage(chatID, "Usage: /set-provider <provider> <model>\nAvailable providers: gemini, kling, recraft\nExample: /set-provider gemini gemini-3.1-flash-image")
			return
		}
		providerName := parts[0]
		modelName := parts[1]

		provider := app.providers.Get(providerName)
		if provider == nil {
			app.bot.SendMessage(chatID, fmt.Sprintf("❌ Provider not found: %s", providerName))
			return
		}

		// Validate model exists.
		found := false
		for _, m := range provider.GetModels() {
			if m == modelName {
				found = true
				break
			}
		}
		if !found {
			app.bot.SendMessage(chatID, fmt.Sprintf("❌ Model not found: %s\nAvailable models: %s", modelName, strings.Join(provider.GetModels(), ", ")))
			return
		}

		if err := app.storage.SetProvider(chatID, providerName, modelName); err != nil {
			slog.Error("failed to set provider", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to save provider.")
			return
		}
		app.bot.SendMessage(chatID, fmt.Sprintf("✅ Provider set to %s with model %s", providerName, modelName))
		return
	}

	// Handle /image
	if strings.HasPrefix(text, "/image") || strings.HasPrefix(caption, "/image") {
		var prompt string
		if strings.HasPrefix(text, "/image") {
			prompt = strings.TrimSpace(strings.TrimPrefix(text, "/image"))
		} else {
			prompt = strings.TrimSpace(strings.TrimPrefix(caption, "/image"))
		}

		if prompt == "" {
			// Get template for this chat
			template, err := app.storage.GetChatTemplate(chatID)
			if err != nil {
				slog.Error("failed to get chat template", "error", err)
			}
			if template != nil && template.ImagePrompt != "" {
				prompt = template.ImagePrompt
			} else {
				app.bot.SendMessage(chatID, "Usage: /image <prompt>\nOr set a template: /templates then /set-template <id>")
				return
			}
		}

		// Download reference image if any.
		var imageData []byte
		var mimeType string

		if message.ReplyToMessage != nil && len(message.ReplyToMessage.Photo) > 0 {
			photo := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
			imageData, _ = app.bot.DownloadFile(photo.FileID)
			mimeType = "image/jpeg"
		} else if len(message.Photo) > 0 {
			photo := message.Photo[len(message.Photo)-1]
			imageData, _ = app.bot.DownloadFile(photo.FileID)
			mimeType = "image/jpeg"
		}

		// Send initial progress message.
		msgID, _ := app.bot.SendMessage(chatID, "🖼️ **Generating image**... *(0s)*")

		job := worker.NewJob(worker.JobTypeImage, prompt, imageData, mimeType, chatID, msgID, func(elapsed int) {
			app.bot.SendProgressMessage(chatID, msgID, "image", elapsed)
		})

		app.pool.Enqueue(job)

		// Handle result asynchronously.
		go func(j worker.Job) {
			result := <-j.ResultChan
			if result.Error != nil {
				slog.Error("image generation failed", "error", result.Error)
				app.bot.SendErrorMessage(chatID, "image")
				return
			}
			app.bot.SendSuccessMessage(chatID, msgID, "image", result.ProviderName, result.ModelName)
			caption := prompt
			if result.ProviderName != "" && result.ModelName != "" {
				caption = fmt.Sprintf("%s\n\n🤖 `%s` | `%s`", prompt, result.ProviderName, result.ModelName)
			}
			if err := app.bot.SendPhoto(chatID, result.Data, caption); err != nil {
				slog.Error("failed to send photo", "error", err)
				app.bot.SendErrorMessage(chatID, "image")
			}
		}(job)

		return
	}

	// Handle /video
	if strings.HasPrefix(text, "/video") || strings.HasPrefix(caption, "/video") {
		var prompt string
		if strings.HasPrefix(text, "/video") {
			prompt = strings.TrimSpace(strings.TrimPrefix(text, "/video"))
		} else {
			prompt = strings.TrimSpace(strings.TrimPrefix(caption, "/video"))
		}

		if prompt == "" {
			// Get template for this chat
			template, err := app.storage.GetChatTemplate(chatID)
			if err != nil {
				slog.Error("failed to get chat template", "error", err)
			}
			if template != nil && template.VideoPrompt != "" {
				prompt = template.VideoPrompt
			} else {
				app.bot.SendMessage(chatID, "Usage: /video <prompt>\nOr set a template: /templates then /set-template <id>")
				return
			}
		}

		// Download reference image if any.
		var imageData []byte
		var mimeType string

		if message.ReplyToMessage != nil && len(message.ReplyToMessage.Photo) > 0 {
			photo := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
			imageData, _ = app.bot.DownloadFile(photo.FileID)
			mimeType = "image/jpeg"
		} else if len(message.Photo) > 0 {
			photo := message.Photo[len(message.Photo)-1]
			imageData, _ = app.bot.DownloadFile(photo.FileID)
			mimeType = "image/jpeg"
		}

		// Send initial progress message.
		msgID, _ := app.bot.SendMessage(chatID, "🎬 **Generating video**... *(0s)* 🎞️")

		job := worker.NewJob(worker.JobTypeVideo, prompt, imageData, mimeType, chatID, msgID, func(elapsed int) {
			app.bot.SendProgressMessage(chatID, msgID, "video", elapsed)
		})

		app.pool.Enqueue(job)

		// Handle result asynchronously.
		go func(j worker.Job) {
			result := <-j.ResultChan
			if result.Error != nil {
				slog.Error("video generation failed", "error", result.Error)
				app.bot.SendErrorMessage(chatID, "video")
				return
			}
			app.bot.SendSuccessMessage(chatID, msgID, "video", result.ProviderName, result.ModelName)
			caption := ""
			if result.ProviderName != "" && result.ModelName != "" {
				caption = fmt.Sprintf("🤖 `%s` | `%s`", result.ProviderName, result.ModelName)
			}
			if err := app.bot.SendVideo(chatID, result.Data, caption); err != nil {
				slog.Error("failed to send video", "error", err)
				app.bot.SendErrorMessage(chatID, "video")
			}
		}(job)

		return
	}
}

// handleJob is the worker pool's job handler.
func (app *App) handleJob(ctx context.Context, job worker.Job) worker.Result {
	slog.Info("handling job", "job_id", job.ID, "type", job.Type)
	start := time.Now()

	// Get provider and model from storage for this chat.
	var providerName, modelName string
	entry, err := app.storage.GetPrompts(job.ChatID)
	if err != nil {
		slog.Error("failed to get prompts for provider selection", "error", err)
	} else if entry != nil {
		providerName = entry.Provider
		modelName = entry.ModelName
	}

	// Fallback to default provider.
	if providerName == "" {
		providerName = app.providers.GetDefault()
	}

	provider := app.providers.Get(providerName)
	if provider == nil {
		result := worker.Result{Error: fmt.Errorf("provider not found: %s", providerName)}
		app.recordGenerationHistory(job, providerName, modelName, result, start)
		return result
	}

	opts := ai.GenerateOptions{
		ModelName: modelName,
	}

	var result worker.Result
	switch job.Type {
	case worker.JobTypeImage:
		data, err := provider.GenerateImage(ctx, job.Prompt, job.ImageData, job.MimeType, opts)
		if err != nil {
			result = worker.Result{Error: err}
		} else {
			result = worker.Result{Data: data, ProviderName: providerName, ModelName: modelName}
		}

	case worker.JobTypeVideo:
		data, err := provider.GenerateVideo(ctx, job.Prompt, job.ImageData, job.MimeType, opts)
		if err != nil {
			result = worker.Result{Error: err}
		} else {
			result = worker.Result{Data: data, ProviderName: providerName, ModelName: modelName}
		}

	default:
		result = worker.Result{Error: fmt.Errorf("unknown job type: %s", job.Type)}
	}

	app.recordGenerationHistory(job, providerName, modelName, result, start)
	return result
}

func (app *App) recordGenerationHistory(job worker.Job, providerName, modelName string, result worker.Result, start time.Time) {
	durationMs := time.Since(start).Milliseconds()
	status := "success"
	resultMap := map[string]any{}
	if result.Error != nil {
		status = "error"
		resultMap["error"] = result.Error.Error()
	} else {
		resultMap["size"] = len(result.Data)
		resultMap["provider"] = result.ProviderName
		resultMap["model"] = result.ModelName
	}

	params := map[string]any{}
	if job.MimeType != "" {
		params["mimeType"] = job.MimeType
	}
	if len(job.ImageData) > 0 {
		params["hasReferenceImage"] = true
		params["referenceImageSize"] = len(job.ImageData)
	}

	_, err := app.storage.CreateGenerationHistory(
		job.ChatID,
		string(job.Type),
		job.Prompt,
		providerName,
		modelName,
		status,
		params,
		resultMap,
		durationMs,
	)
	if err != nil {
		slog.Error("failed to record generation history", "error", err)
	}
}
