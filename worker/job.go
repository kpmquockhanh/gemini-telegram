package worker

import "context"

// JobType represents the type of AI generation job.
type JobType string

const (
	JobTypeImage JobType = "image"
	JobTypeVideo JobType = "video"
)

// Job represents a single unit of work for the worker pool.
type Job struct {
	ID        int64
	Type      JobType
	Prompt    string
	ImageData []byte
	MimeType  string
	ChatID    int64
	MessageID int

	// ProgressFn is called periodically with elapsed seconds.
	// It runs in its own goroutine and should be non-blocking.
	ProgressFn func(elapsed int)

	// ResultChan receives the final result.
	ResultChan chan Result

	// Ctx allows per-job cancellation.
	Ctx context.Context
}

// Result holds the outcome of a job.
type Result struct {
	Data  []byte
	Error error
}

// NewJob creates a new Job with the given parameters.
func NewJob(jobType JobType, prompt string, imageData []byte, mimeType string, chatID int64, messageID int, progressFn func(elapsed int)) Job {
	return Job{
		ID:         0, // Assigned by the pool
		Type:       jobType,
		Prompt:     prompt,
		ImageData:  imageData,
		MimeType:   mimeType,
		ChatID:     chatID,
		MessageID:  messageID,
		ProgressFn: progressFn,
		ResultChan: make(chan Result, 1),
		Ctx:        context.Background(),
	}
}
