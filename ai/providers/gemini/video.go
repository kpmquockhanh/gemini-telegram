package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"gemini-telegram-bot/ai"
	"google.golang.org/genai"
)

// GenerateVideo creates a video using Veo.
func (c *Client) GenerateVideo(ctx context.Context, prompt string, imageData []byte, mimeType string, opts ai.GenerateOptions) ([]byte, error) {
	source := &genai.GenerateVideosSource{
		Prompt: prompt,
	}

	if imageData != nil && len(imageData) > 0 {
		if mimeType == "" {
			mimeType = "image/png"
		}
		source.Image = &genai.Image{
			ImageBytes: imageData,
			MIMEType:   mimeType,
		}
	}

	config := &genai.GenerateVideosConfig{
		DurationSeconds: genai.Ptr(durationFromOpts(opts)),
	}

	modelName := opts.ModelName
	if modelName == "" {
		modelName = "veo-3.1-generate-preview"
	}

	slog.Info("starting video generation", "prompt", prompt, "model", modelName)
	operation, err := c.client.Models.GenerateVideosFromSource(ctx, modelName, source, config)
	if err != nil {
		return nil, fmt.Errorf("generate videos failed: %w", err)
	}

	// Poll until the operation is done.
	for !operation.Done {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("video generation cancelled")
		case <-time.After(10 * time.Second):
		}

		operation, err = c.client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			return nil, fmt.Errorf("polling operation failed: %w", err)
		}
		slog.Info("video generation polling", "done", operation.Done)
	}

	if operation.Response == nil || len(operation.Response.GeneratedVideos) == 0 {
		return nil, fmt.Errorf("no videos generated")
	}

	video := operation.Response.GeneratedVideos[0].Video
	if video == nil {
		return nil, fmt.Errorf("generated video is nil")
	}

	slog.Info("video generation completed, downloading...")

	// Download the video using the Files API.
	data, err := c.client.Files.Download(ctx, video, nil)
	if err != nil {
		return nil, fmt.Errorf("downloading video failed: %w", err)
	}

	slog.Info("video downloaded successfully", "size", len(data))
	return data, nil
}

func durationFromOpts(opts ai.GenerateOptions) int32 {
	if opts.Params == nil {
		return 4
	}
	v, ok := opts.Params["duration"]
	if !ok {
		return 4
	}
	switch val := v.(type) {
	case float64:
		return int32(val)
	case float32:
		return int32(val)
	case int:
		return int32(val)
	case int32:
		return val
	case json.Number:
		i, _ := val.Int64()
		if i >= 4 && i <= 8 {
			return int32(i)
		}
		return 4
	default:
		return 4
	}
}
