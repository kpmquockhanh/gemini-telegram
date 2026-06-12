package gemini

import (
	"context"
	"log/slog"

	"google.golang.org/genai"
)

// Client wraps the Google GenAI SDK.
type Client struct {
	client *genai.Client
}

// NewClient creates a new Gemini client with the given API key.
func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	cfg := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}

	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	slog.Info("gemini client initialized")
	return &Client{client: client}, nil
}
