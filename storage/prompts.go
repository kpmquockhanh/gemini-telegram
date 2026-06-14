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
	ChatID       int64  `json:"chatId"`
	ImagePrompt  string `json:"imagePrompt"`
	VideoPrompt  string `json:"videoPrompt"`
	Provider     string `json:"provider"`
	ModelName    string `json:"modelName"`
	UpdatedAt    string `json:"updatedAt"`
}

// TemplateEntry holds a reusable prompt template.
type TemplateEntry struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImagePrompt string `json:"imagePrompt"`
	VideoPrompt string `json:"videoPrompt"`
	CreatedAt   string `json:"createdAt"`
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

// 	migrate creates the required tables.
func (s *PromptStore) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS default_prompts (
    chat_id INTEGER PRIMARY KEY,
    image_prompt TEXT,
    video_prompt TEXT,
    provider TEXT,
    model_name TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS prompt_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    image_prompt TEXT,
    video_prompt TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Migrate: add provider and model_name columns if they don't exist.
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN provider TEXT`)
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN model_name TEXT`)

	return nil
}

// GetPrompts retrieves the default prompts for a chat.
func (s *PromptStore) GetPrompts(chatID int64) (*PromptEntry, error) {
	row := s.db.QueryRow(
		`SELECT chat_id, image_prompt, video_prompt, provider, model_name, updated_at FROM default_prompts WHERE chat_id = ?`,
		chatID,
	)

	var entry PromptEntry
	err := row.Scan(&entry.ChatID, &entry.ImagePrompt, &entry.VideoPrompt, &entry.Provider, &entry.ModelName, &entry.UpdatedAt)
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

// SetProvider sets the AI provider and model for a chat.
func (s *PromptStore) SetProvider(chatID int64, provider, modelName string) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, provider, model_name, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			provider = excluded.provider,
			model_name = excluded.model_name,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, provider, modelName)
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

// ListAllPrompts returns a paginated list of prompts with total count.
func (s *PromptStore) ListAllPrompts(offset, limit int, search string) ([]PromptEntry, int, error) {
	// Build query with optional search.
	whereClause := ""
	args := []any{}
	if search != "" {
		whereClause = "WHERE chat_id LIKE ? OR image_prompt LIKE ? OR video_prompt LIKE ?"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	// Count total.
	var total int
	countQuery := "SELECT COUNT(*) FROM default_prompts " + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch rows.
	query := "SELECT chat_id, image_prompt, video_prompt, provider, model_name, updated_at FROM default_prompts " + whereClause + " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []PromptEntry
	for rows.Next() {
		var entry PromptEntry
		if err := rows.Scan(&entry.ChatID, &entry.ImagePrompt, &entry.VideoPrompt, &entry.Provider, &entry.ModelName, &entry.UpdatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

// UpdatePrompts upserts both image and video prompts for a chat.
func (s *PromptStore) UpdatePrompts(chatID int64, imagePrompt, videoPrompt, provider, modelName string) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, image_prompt, video_prompt, provider, model_name, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			image_prompt = excluded.image_prompt,
			video_prompt = excluded.video_prompt,
			provider = excluded.provider,
			model_name = excluded.model_name,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, imagePrompt, videoPrompt, provider, modelName)
	return err
}

// DeletePrompts removes a chat's prompt entry.
func (s *PromptStore) DeletePrompts(chatID int64) error {
	_, err := s.db.Exec(`DELETE FROM default_prompts WHERE chat_id = ?`, chatID)
	return err
}

// GetStats returns dashboard statistics.
func (s *PromptStore) GetStats() (totalChats, imageCount, videoCount, totalPrompts int, err error) {
	err = s.db.QueryRow(`
		SELECT 
			COUNT(*),
			IFNULL(SUM(CASE WHEN image_prompt IS NOT NULL AND image_prompt != '' THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN video_prompt IS NOT NULL AND video_prompt != '' THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN (image_prompt IS NOT NULL AND image_prompt != '') OR (video_prompt IS NOT NULL AND video_prompt != '') THEN 1 ELSE 0 END), 0)
		FROM default_prompts
	`).Scan(&totalChats, &imageCount, &videoCount, &totalPrompts)
	return
}

// ListTemplates returns all prompt templates.
func (s *PromptStore) ListTemplates() ([]TemplateEntry, error) {
	rows, err := s.db.Query(`SELECT id, name, description, image_prompt, video_prompt, created_at FROM prompt_templates ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TemplateEntry
	for rows.Next() {
		var entry TemplateEntry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Description, &entry.ImagePrompt, &entry.VideoPrompt, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// GetTemplate retrieves a single template by ID.
func (s *PromptStore) GetTemplate(id int64) (*TemplateEntry, error) {
	row := s.db.QueryRow(`SELECT id, name, description, image_prompt, video_prompt, created_at FROM prompt_templates WHERE id = ?`, id)
	var entry TemplateEntry
	err := row.Scan(&entry.ID, &entry.Name, &entry.Description, &entry.ImagePrompt, &entry.VideoPrompt, &entry.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// CreateTemplate adds a new template.
func (s *PromptStore) CreateTemplate(name, description, imagePrompt, videoPrompt string) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO prompt_templates (name, description, image_prompt, video_prompt)
		VALUES (?, ?, ?, ?)
	`, name, description, imagePrompt, videoPrompt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateTemplate modifies an existing template.
func (s *PromptStore) UpdateTemplate(id int64, name, description, imagePrompt, videoPrompt string) error {
	_, err := s.db.Exec(`
		UPDATE prompt_templates SET
			name = ?,
			description = ?,
			image_prompt = ?,
			video_prompt = ?
		WHERE id = ?
	`, name, description, imagePrompt, videoPrompt, id)
	return err
}

// DeleteTemplate removes a template.
func (s *PromptStore) DeleteTemplate(id int64) error {
	_, err := s.db.Exec(`DELETE FROM prompt_templates WHERE id = ?`, id)
	return err
}

// Close closes the database connection.
func (s *PromptStore) Close() error {
	return s.db.Close()
}
