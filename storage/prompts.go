package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

// PromptStore persists default prompts per chat.
type PromptStore struct {
	db *sql.DB
}

// PromptEntry holds the default prompts for a single chat.
type PromptEntry struct {
	ChatID       int64
	ImagePrompt  string
	VideoPrompt  string
}

// NewPromptStore opens (or creates) the SQLite database at the given path.
func NewPromptStore(dbPath string) (*PromptStore, error) {
	// Ensure the directory exists.
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging sqlite db: %w", err)
	}

	store := &PromptStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return store, nil
}

// migrate creates the required tables.
func (s *PromptStore) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS default_prompts (
    chat_id INTEGER PRIMARY KEY,
    image_prompt TEXT,
    video_prompt TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
	_, err := s.db.Exec(schema)
	return err
}

// GetPrompts retrieves the default prompts for a chat.
func (s *PromptStore) GetPrompts(chatID int64) (*PromptEntry, error) {
	row := s.db.QueryRow(
		`SELECT chat_id, image_prompt, video_prompt FROM default_prompts WHERE chat_id = ?`,
		chatID,
	)

	var entry PromptEntry
	err := row.Scan(&entry.ChatID, &entry.ImagePrompt, &entry.VideoPrompt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// SetImagePrompt sets the default image prompt for a chat.
func (s *PromptStore) SetImagePrompt(chatID int64, prompt string) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, image_prompt, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			image_prompt = excluded.image_prompt,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, prompt)
	return err
}

// SetVideoPrompt sets the default video prompt for a chat.
func (s *PromptStore) SetVideoPrompt(chatID int64, prompt string) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, video_prompt, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			video_prompt = excluded.video_prompt,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, prompt)
	return err
}

// Close closes the database connection.
func (s *PromptStore) Close() error {
	return s.db.Close()
}
