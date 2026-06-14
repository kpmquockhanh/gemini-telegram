package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gemini-telegram-bot/ai"
	"gemini-telegram-bot/ai/providers/gemini"
	"gemini-telegram-bot/ai/providers/kling"
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

	// Handle /set-image-prompt
	if strings.HasPrefix(text, "/set-image-prompt") {
		prompt := strings.TrimSpace(strings.TrimPrefix(text, "/set-image-prompt"))
		if prompt == "" {
			app.bot.SendMessage(chatID, "Usage: /set-image-prompt <prompt>\nSets the default prompt used when /image is sent without one.")
			return
		}
		if err := app.storage.SetImagePrompt(chatID, prompt); err != nil {
			slog.Error("failed to set image prompt", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to save default prompt.")
			return
		}
		app.bot.SendMessage(chatID, fmt.Sprintf("✅ Default image prompt set to:\n\"%s\"", prompt))
		return
	}

	// Handle /set-video-prompt
	if strings.HasPrefix(text, "/set-video-prompt") {
		prompt := strings.TrimSpace(strings.TrimPrefix(text, "/set-video-prompt"))
		if prompt == "" {
			app.bot.SendMessage(chatID, "Usage: /set-video-prompt <prompt>\nSets the default prompt used when /video is sent without one.")
			return
		}
		if err := app.storage.SetVideoPrompt(chatID, prompt); err != nil {
			slog.Error("failed to set video prompt", "error", err)
			app.bot.SendMessage(chatID, "❌ Failed to save default prompt.")
			return
		}
		app.bot.SendMessage(chatID, fmt.Sprintf("✅ Default video prompt set to:\n\"%s\"", prompt))
		return
	}

	// Handle /set-provider
	if strings.HasPrefix(text, "/set-provider") {
		args := strings.TrimSpace(strings.TrimPrefix(text, "/set-provider"))
		parts := strings.Fields(args)
		if len(parts) < 2 {
			app.bot.SendMessage(chatID, "Usage: /set-provider <provider> <model>\nAvailable providers: gemini, kling\nExample: /set-provider gemini gemini-3.1-flash-image")
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
			entry, err := app.storage.GetPrompts(chatID)
			if err != nil {
				slog.Error("failed to get prompts", "error", err)
			}
			if entry != nil && entry.ImagePrompt != "" {
				prompt = entry.ImagePrompt
			} else {
				app.bot.SendMessage(chatID, "Usage: /image <prompt>\nOr set a default: /set-image-prompt <prompt>")
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
			app.bot.SendSuccessMessage(chatID, msgID, "image")
			if err := app.bot.SendPhoto(chatID, result.Data, prompt); err != nil {
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
			entry, err := app.storage.GetPrompts(chatID)
			if err != nil {
				slog.Error("failed to get prompts", "error", err)
			}
			if entry != nil && entry.VideoPrompt != "" {
				prompt = entry.VideoPrompt
			} else {
				app.bot.SendMessage(chatID, "Usage: /video <prompt>\nOr set a default: /set-video-prompt <prompt>")
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
			app.bot.SendSuccessMessage(chatID, msgID, "video")
			if err := app.bot.SendVideo(chatID, result.Data); err != nil {
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
		return worker.Result{Error: fmt.Errorf("provider not found: %s", providerName)}
	}

	opts := ai.GenerateOptions{
		ModelName: modelName,
	}

	switch job.Type {
	case worker.JobTypeImage:
		data, err := provider.GenerateImage(ctx, job.Prompt, job.ImageData, job.MimeType, opts)
		if err != nil {
			return worker.Result{Error: err}
		}
		return worker.Result{Data: data}

	case worker.JobTypeVideo:
		data, err := provider.GenerateVideo(ctx, job.Prompt, job.ImageData, job.MimeType, opts)
		if err != nil {
			return worker.Result{Error: err}
		}
		return worker.Result{Data: data}

	default:
		return worker.Result{Error: fmt.Errorf("unknown job type: %s", job.Type)}
	}
}
