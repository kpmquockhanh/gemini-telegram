package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

// PromptStore persists default prompts per chat.
type PromptStore struct {
	db *sql.DB
}

// PromptEntry holds the default prompt settings for a single chat.
// Prompts are resolved through the template_id.
type PromptEntry struct {
	ChatID       int64  `json:"chatId"`
	TemplateID   *int64 `json:"templateId"`
	TemplateName string `json:"templateName"`
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

// GenerationHistoryEntry holds a record of an image/video generation request.
type GenerationHistoryEntry struct {
	ID         int64  `json:"id"`
	ChatID     int64  `json:"chatId"`
	JobType    string `json:"jobType"`
	Prompt     string `json:"prompt"`
	Provider   string `json:"provider"`
	ModelName  string `json:"modelName"`
	Status     string `json:"status"`
	Params     string `json:"params"`
	Result     string `json:"result"`
	DurationMs int64  `json:"durationMs"`
	CreatedAt  string `json:"createdAt"`
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

// migrate creates the required tables and runs migrations.
func (s *PromptStore) migrate() error {
	// Create tables with new schema
	schema := `
CREATE TABLE IF NOT EXISTS default_prompts (
    chat_id INTEGER PRIMARY KEY,
    template_id INTEGER,
    provider TEXT,
    model_name TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES prompt_templates(id)
);
CREATE TABLE IF NOT EXISTS prompt_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    image_prompt TEXT,
    video_prompt TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS generation_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER,
    job_type TEXT NOT NULL,
    prompt TEXT,
    provider TEXT,
    model_name TEXT,
    status TEXT NOT NULL,
    params TEXT,
    result TEXT,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_generation_history_created_at ON generation_history(created_at);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Migrate old schema: add columns that might not exist
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN template_id INTEGER`)
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN provider TEXT`)
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN model_name TEXT`)

	// Migrate old columns to new template-based system
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN image_prompt TEXT`)
	_, _ = s.db.Exec(`ALTER TABLE default_prompts ADD COLUMN video_prompt TEXT`)

	// Run data migration: convert existing prompts to templates
	if err := s.migrateOldPrompts(); err != nil {
		return fmt.Errorf("migrating old prompts: %w", err)
	}

	return nil
}

// migrateOldPrompts converts old image_prompt/video_prompt fields into templates.
func (s *PromptStore) migrateOldPrompts() error {
	// Check if there are old-style prompts with image_prompt or video_prompt but no template_id
	rows, err := s.db.Query(`
		SELECT chat_id, image_prompt, video_prompt, provider, model_name 
		FROM default_prompts 
		WHERE (image_prompt IS NOT NULL OR video_prompt IS NOT NULL) 
		AND template_id IS NULL
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var chatID int64
		var imagePrompt, videoPrompt, provider, modelName sql.NullString
		if err := rows.Scan(&chatID, &imagePrompt, &videoPrompt, &provider, &modelName); err != nil {
			continue
		}

		// Create a template from the old prompts
		name := fmt.Sprintf("Auto-migrated (Chat %d)", chatID)
		res, err := s.db.Exec(`
			INSERT INTO prompt_templates (name, description, image_prompt, video_prompt)
			VALUES (?, ?, ?, ?)
		`, name, "Auto-migrated from old prompt system", imagePrompt.String, videoPrompt.String)
		if err != nil {
			continue
		}

		templateID, err := res.LastInsertId()
		if err != nil {
			continue
		}

		// Update the chat to reference the new template
		_, _ = s.db.Exec(`
			UPDATE default_prompts 
			SET template_id = ?, provider = ?, model_name = ?
			WHERE chat_id = ?
		`, templateID, provider.String, modelName.String, chatID)
	}

	// Drop old columns after migration (sqlite doesn't support DROP COLUMN easily, so we just ignore them)
	// We keep them but don't use them in the new code.
	return nil
}

// GetPrompts retrieves the prompt settings for a chat (with resolved template name).
func (s *PromptStore) GetPrompts(chatID int64) (*PromptEntry, error) {
	row := s.db.QueryRow(`
		SELECT d.chat_id, d.template_id, t.name, d.provider, d.model_name, d.updated_at
		FROM default_prompts d
		LEFT JOIN prompt_templates t ON d.template_id = t.id
		WHERE d.chat_id = ?
	`, chatID)

	var entry PromptEntry
	var templateID sql.NullInt64
	var templateName sql.NullString
	var provider sql.NullString
	var modelName sql.NullString
	err := row.Scan(&entry.ChatID, &templateID, &templateName, &provider, &modelName, &entry.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if templateID.Valid {
		id := templateID.Int64
		entry.TemplateID = &id
		entry.TemplateName = templateName.String
	}
	if provider.Valid {
		entry.Provider = provider.String
	}
	if modelName.Valid {
		entry.ModelName = modelName.String
	}
	return &entry, nil
}

// GetChatTemplate retrieves the template assigned to a chat.
func (s *PromptStore) GetChatTemplate(chatID int64) (*TemplateEntry, error) {
	row := s.db.QueryRow(`
		SELECT t.id, t.name, t.description, t.image_prompt, t.video_prompt, t.created_at
		FROM prompt_templates t
		JOIN default_prompts d ON t.id = d.template_id
		WHERE d.chat_id = ?
	`, chatID)

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

// SetTemplate assigns a template to a chat.
func (s *PromptStore) SetTemplate(chatID int64, templateID int64) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, template_id, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			template_id = excluded.template_id,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, templateID)
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

// ListAllPrompts returns a paginated list of prompts with total count.
func (s *PromptStore) ListAllPrompts(offset, limit int, search string) ([]PromptEntry, int, error) {
	// Build query with optional search.
	whereClause := ""
	args := []any{}
	if search != "" {
		whereClause = "WHERE d.chat_id LIKE ? OR t.name LIKE ?"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	// Count total.
	var total int
	countQuery := "SELECT COUNT(*) FROM default_prompts d LEFT JOIN prompt_templates t ON d.template_id = t.id " + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch rows.
	query := `
		SELECT d.chat_id, d.template_id, t.name, d.provider, d.model_name, d.updated_at 
		FROM default_prompts d
		LEFT JOIN prompt_templates t ON d.template_id = t.id
	` + whereClause + " ORDER BY d.updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []PromptEntry
	for rows.Next() {
		var entry PromptEntry
		var templateID sql.NullInt64
		var templateName sql.NullString
		var provider sql.NullString
		var modelName sql.NullString
		if err := rows.Scan(&entry.ChatID, &templateID, &templateName, &provider, &modelName, &entry.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if templateID.Valid {
			id := templateID.Int64
			entry.TemplateID = &id
			entry.TemplateName = templateName.String
		}
		if provider.Valid {
			entry.Provider = provider.String
		}
		if modelName.Valid {
			entry.ModelName = modelName.String
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

// UpdatePrompts updates the template assignment and provider for a chat.
func (s *PromptStore) UpdatePrompts(chatID int64, templateID int64, provider, modelName string) error {
	_, err := s.db.Exec(`
		INSERT INTO default_prompts (chat_id, template_id, provider, model_name, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_id) DO UPDATE SET
			template_id = excluded.template_id,
			provider = excluded.provider,
			model_name = excluded.model_name,
			updated_at = CURRENT_TIMESTAMP
	`, chatID, templateID, provider, modelName)
	return err
}

// DeletePrompts removes a chat's prompt entry.
func (s *PromptStore) DeletePrompts(chatID int64) error {
	_, err := s.db.Exec(`DELETE FROM default_prompts WHERE chat_id = ?`, chatID)
	return err
}

// GetStats returns dashboard statistics.
func (s *PromptStore) GetStats() (totalChats, templateCount, totalPrompts int, err error) {
	err = s.db.QueryRow(`
		SELECT 
			COUNT(*),
			IFNULL(SUM(CASE WHEN template_id IS NOT NULL THEN 1 ELSE 0 END), 0),
			COUNT(*)
		FROM default_prompts
	`).Scan(&totalChats, &templateCount, &totalPrompts)
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

// CreateGenerationHistory adds a new generation history record.
func (s *PromptStore) CreateGenerationHistory(chatID int64, jobType, prompt, provider, modelName, status string, params, result map[string]any, durationMs int64) (int64, error) {
	var paramsJSON, resultJSON []byte
	var err error

	if params != nil {
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			paramsJSON = []byte("{}")
		}
	}
	if result != nil {
		resultJSON, err = json.Marshal(result)
		if err != nil {
			resultJSON = []byte("{}")
		}
	}

	res, err := s.db.Exec(`
		INSERT INTO generation_history (chat_id, job_type, prompt, provider, model_name, status, params, result, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, chatID, jobType, prompt, provider, modelName, status, string(paramsJSON), string(resultJSON), durationMs)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListGenerationHistory returns a paginated list of generation history entries.
func (s *PromptStore) ListGenerationHistory(offset, limit int, statusFilter, search string) ([]GenerationHistoryEntry, int, error) {
	whereClause := ""
	args := []any{}

	if statusFilter != "" && statusFilter != "all" {
		whereClause = "WHERE status = ?"
		args = append(args, statusFilter)
	}
	if search != "" {
		if whereClause != "" {
			whereClause += " AND (prompt LIKE ? OR provider LIKE ? OR model_name LIKE ?)"
		} else {
			whereClause = "WHERE prompt LIKE ? OR provider LIKE ? OR model_name LIKE ?"
		}
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM generation_history " + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, chat_id, job_type, prompt, provider, model_name, status, params, result, duration_ms, created_at
		FROM generation_history
	` + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []GenerationHistoryEntry
	for rows.Next() {
		var entry GenerationHistoryEntry
		var provider, modelName, paramsStr, resultStr sql.NullString
		var durationMs sql.NullInt64
		if err := rows.Scan(&entry.ID, &entry.ChatID, &entry.JobType, &entry.Prompt, &provider, &modelName, &entry.Status, &paramsStr, &resultStr, &durationMs, &entry.CreatedAt); err != nil {
			return nil, 0, err
		}
		if provider.Valid {
			entry.Provider = provider.String
		}
		if modelName.Valid {
			entry.ModelName = modelName.String
		}
		if paramsStr.Valid {
			entry.Params = paramsStr.String
		}
		if resultStr.Valid {
			entry.Result = resultStr.String
		}
		if durationMs.Valid {
			entry.DurationMs = durationMs.Int64
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

// ClearGenerationHistory removes all generation history records.
func (s *PromptStore) ClearGenerationHistory() error {
	_, err := s.db.Exec(`DELETE FROM generation_history`)
	return err
}

// Close closes the database connection.
func (s *PromptStore) Close() error {
	return s.db.Close()
}
