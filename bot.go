package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// BotClient wraps the Telegram Bot API.
type BotClient struct {
	token string
	client *http.Client
}

// NewBotClient creates a new Telegram bot client.
func NewBotClient(token string) *BotClient {
	return &BotClient{
		token: token,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (b *BotClient) apiURL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}

// SendMessage sends a text message to a chat.
func (b *BotClient) SendMessage(chatID int64, text string) (int, error) {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)

	resp, err := b.client.Post(b.apiURL("sendMessage"), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}
	_ = json.Unmarshal(respBody, &result)

	if !result.OK {
		return 0, fmt.Errorf("telegram API error: %s", string(respBody))
	}
	return result.Result.MessageID, nil
}

// EditMessageText edits an existing message.
func (b *BotClient) EditMessageText(chatID int64, messageID int, text string) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)

	resp, err := b.client.Post(b.apiURL("editMessageText"), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(respBody, &result)

	if !result.OK {
		// Ignore "message not modified" errors.
		if bytes.Contains(respBody, []byte("message is not modified")) {
			return nil
		}
		return fmt.Errorf("telegram API error: %s", string(respBody))
	}
	return nil
}

// SendPhoto sends a photo to a chat.
func (b *BotClient) SendPhoto(chatID int64, imageData []byte, caption string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	part, err := writer.CreateFormFile("photo", "image.png")
	if err != nil {
		return err
	}
	_, err = part.Write(imageData)
	if err != nil {
		return err
	}
	writer.Close()

	resp, err := b.client.Post(b.apiURL("sendPhoto"), writer.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(respBody, &result)

	if !result.OK {
		return fmt.Errorf("telegram API error: %s", string(respBody))
	}
	return nil
}

// SendVideo sends a video to a chat.
func (b *BotClient) SendVideo(chatID int64, videoData []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))

	part, err := writer.CreateFormFile("video", "video.mp4")
	if err != nil {
		return err
	}
	_, err = part.Write(videoData)
	if err != nil {
		return err
	}
	writer.Close()

	resp, err := b.client.Post(b.apiURL("sendVideo"), writer.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(respBody, &result)

	if !result.OK {
		return fmt.Errorf("telegram API error: %s", string(respBody))
	}
	return nil
}

// DownloadFile downloads a file from Telegram servers.
func (b *BotClient) DownloadFile(fileID string) ([]byte, error) {
	// Get file path.
	resp, err := b.client.Get(b.apiURL("getFile") + "?file_id=" + fileID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	_ = json.Unmarshal(respBody, &result)

	if !result.OK || result.Result.FilePath == "" {
		return nil, fmt.Errorf("failed to get file path: %s", string(respBody))
	}

	// Download the file.
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.token, result.Result.FilePath)
	resp, err = b.client.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// SendProgressMessage sends or edits a progress message.
func (b *BotClient) SendProgressMessage(chatID int64, messageID int, jobType string, elapsed int) error {
	var icon string
	var noun string
	if jobType == "image" {
		icon = "🖼️"
		noun = "image"
	} else {
		icon = "🎬"
		noun = "video"
	}

	text := fmt.Sprintf("%s **Generating %s**... *(%ds)* ⏳", icon, noun, elapsed)

	if messageID == 0 {
		id, err := b.SendMessage(chatID, text)
		if err != nil {
			return err
		}
		slog.Info("sent progress message", "chat_id", chatID, "message_id", id)
		return nil
	}

	return b.EditMessageText(chatID, messageID, text)
}

// SendErrorMessage sends an error message to a chat.
func (b *BotClient) SendErrorMessage(chatID int64, jobType string) error {
	var noun string
	if jobType == "image" {
		noun = "image"
	} else {
		noun = "video"
	}

	text := fmt.Sprintf("❌ **Failed to generate %s**\n\nPlease try again later or check your prompt.", noun)
	_, err := b.SendMessage(chatID, text)
	return err
}

// SendSuccessMessage sends a success placeholder (used before sending media).
func (b *BotClient) SendSuccessMessage(chatID int64, messageID int, jobType string) error {
	var icon string
	var noun string
	if jobType == "image" {
		icon = "✅"
		noun = "Image"
	} else {
		icon = "✅"
		noun = "Video"
	}

	text := fmt.Sprintf("%s **%s generated!** Sending now...", icon, noun)
	return b.EditMessageText(chatID, messageID, text)
}
