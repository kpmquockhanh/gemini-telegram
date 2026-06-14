package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"gemini-telegram-bot/ai"
	"google.golang.org/genai"
)

// GenerateImage creates an image using Gemini Flash.
func (c *Client) GenerateImage(ctx context.Context, prompt string, imageData []byte, mimeType string, opts ai.GenerateOptions) ([]byte, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	if imageData != nil && len(imageData) > 0 {
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		contents = append(contents, genai.NewContentFromBytes(imageData, mimeType, "user"))
	}

	modelName := opts.ModelName
	if modelName == "" {
		modelName = "gemini-3.1-flash-image"
	}

	resp, err := c.client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("generate content failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response")
	}

	for _, part := range candidate.Content.Parts {
		if part.InlineData != nil {
			data := part.InlineData.Data
			slog.Info("image generated successfully", "size", len(data))
			return data, nil
		}
	}

	return nil, fmt.Errorf("no inline data found in response")
}
