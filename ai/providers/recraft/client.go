package recraft

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

const baseURL = "https://external.api.recraft.ai/v1"

// Client wraps the Recraft AI API.
type Client struct {
	apiKey string
	http   *http.Client
}

// NewClient creates a new Recraft AI client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

// Name returns the provider identifier.
func (c *Client) Name() string {
	return "recraft"
}

// GetModels returns available Recraft AI models.
func (c *Client) GetModels() []string {
	return []string{
		"recraftv4_1",
		"recraftv4_1_vector",
		"recraftv4_1_pro",
		"recraftv4_1_pro_vector",
		"recraftv4_1_utility",
		"recraftv4_1_utility_vector",
		"recraftv4_1_utility_pro",
		"recraftv4_1_utility_pro_vector",
		"recraftv4",
		"recraftv4_vector",
		"recraftv4_pro",
		"recraftv4_pro_vector",
		"recraftv3",
		"recraftv3_vector",
		"recraftv2",
		"recraftv2_vector",
	}
}

// doRequest performs an HTTP request with Bearer auth and returns the response body.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("recraft request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recraft returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("recraft read body failed: %w", err)
	}

	return body, nil
}

// generationResponse is the JSON shape returned by Recraft image endpoints.
type generationResponse struct {
	Data []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

// parseImageData extracts the first image from a Recraft JSON response.
func parseImageData(raw []byte) ([]byte, error) {
	var resp generationResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("recraft unmarshal failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("recraft returned no images")
	}

	// Prefer b64_json if present.
	if resp.Data[0].B64JSON != "" {
		slog.Info("recraft image generated (b64_json)", "size", len(resp.Data[0].B64JSON))
		return nil, fmt.Errorf("recraft returned base64 data but decoder not yet wired")
	}

	if resp.Data[0].URL != "" {
		return nil, fmt.Errorf("recraft returned URL but direct download not yet wired: %s", resp.Data[0].URL)
	}

	return nil, fmt.Errorf("recraft response contained no image data")
}

// parseImageDataB64 extracts the first base64-encoded image from a Recraft JSON response.
func parseImageDataB64(raw []byte) ([]byte, error) {
	var resp generationResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("recraft unmarshal failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("recraft returned no images")
	}

	b64 := resp.Data[0].B64JSON
	if b64 == "" {
		return nil, fmt.Errorf("recraft response contained no b64_json data")
	}

	return []byte(b64), nil
}

// writeMultipartField writes a simple string field to a multipart writer.
func writeMultipartField(w *multipart.Writer, fieldName, value string) {
	_ = w.WriteField(fieldName, value)
}
