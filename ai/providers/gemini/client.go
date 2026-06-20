package gemini

import (
	"context"
	"log/slog"

	"gemini-telegram-bot/ai"
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

// Name returns the provider identifier.
func (c *Client) Name() string {
	return "gemini"
}

// GetModels returns available Gemini models.
func (c *Client) GetModels() []string {
	return []string{
		"gemini-3.1-flash-image",
		"veo-3.1-generate-preview",
	}
}

// GetModelParams returns configurable parameters for a Gemini model.
func (c *Client) GetModelParams(modelName string) []ai.ModelParamDef {
	switch modelName {
	case "veo-3.1-generate-preview":
		min := 4.0
		max := 8.0
		step := 1.0
		return []ai.ModelParamDef{
			{
				Name:    "duration",
				Label:   "Duration (seconds)",
				Type:    "number",
				Default: 4,
				Min:     &min,
				Max:     &max,
				Step:    &step,
			},
		}
	default:
		return nil
	}
}
