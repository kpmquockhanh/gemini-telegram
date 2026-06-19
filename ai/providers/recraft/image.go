package recraft

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"

	"gemini-telegram-bot/ai"
)

// GenerateImage creates an image using Recraft AI.
// If imageData is provided, it uses the image-to-image endpoint; otherwise text-to-image.
func (c *Client) GenerateImage(ctx context.Context, prompt string, imageData []byte, mimeType string, opts ai.GenerateOptions) ([]byte, error) {
	if imageData != nil && len(imageData) > 0 {
		return c.generateImageToImage(ctx, prompt, imageData, mimeType, opts)
	}
	return c.generateTextToImage(ctx, prompt, opts)
}

// GenerateVideo returns an error because Recraft does not support video generation.
func (c *Client) GenerateVideo(ctx context.Context, prompt string, imageData []byte, mimeType string, opts ai.GenerateOptions) ([]byte, error) {
	return nil, fmt.Errorf("recraft does not support video generation")
}

// generateTextToImage calls POST /images/generations.
func (c *Client) generateTextToImage(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]byte, error) {
	modelName := opts.ModelName
	if modelName == "" {
		modelName = "recraftv4_1"
	}

	slog.Info("generating image with recraft", "model", modelName, "prompt", prompt)

	body := map[string]any{
		"prompt":          prompt,
		"model":           modelName,
		"n":               1,
		"response_format": "b64_json",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("recraft marshal failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("recraft create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	raw, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	b64, err := parseImageDataB64(raw)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		return nil, fmt.Errorf("recraft base64 decode failed: %w", err)
	}

	slog.Info("recraft image generated successfully", "size", len(data))
	return data, nil
}

// generateImageToImage calls POST /images/imageToImage with multipart form data.
func (c *Client) generateImageToImage(ctx context.Context, prompt string, imageData []byte, mimeType string, opts ai.GenerateOptions) ([]byte, error) {
	modelName := opts.ModelName
	if modelName == "" {
		modelName = "recraftv3"
	}

	// imageToImage only supports recraftv3 / recraftv3_vector.
	if modelName != "recraftv3" && modelName != "recraftv3_vector" {
		modelName = "recraftv3"
	}

	slog.Info("generating image-to-image with recraft", "model", modelName, "prompt", prompt)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Write image file.
	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		return nil, fmt.Errorf("recraft create form file failed: %w", err)
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("recraft write image data failed: %w", err)
	}

	writeMultipartField(writer, "prompt", prompt)
	writeMultipartField(writer, "strength", "0.5")
	writeMultipartField(writer, "model", modelName)
	writeMultipartField(writer, "response_format", "b64_json")

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("recraft close multipart failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/images/imageToImage", &buf)
	if err != nil {
		return nil, fmt.Errorf("recraft create request failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	raw, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	b64, err := parseImageDataB64(raw)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		return nil, fmt.Errorf("recraft base64 decode failed: %w", err)
	}

	slog.Info("recraft image-to-image generated successfully", "size", len(data))
	return data, nil
}
