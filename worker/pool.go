package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Pool manages a fixed number of worker goroutines.
type Pool struct {
	size      int
	jobs      chan Job
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	handler   JobHandler
	jobCounter int64
}

// JobHandler processes a single job.
type JobHandler func(ctx context.Context, job Job) Result

// NewPool creates a worker pool with the given size and job handler.
func NewPool(size int, handler JobHandler) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &Pool{
		size:    size,
		jobs:    make(chan Job, size),
		ctx:     ctx,
		cancel:  cancel,
		handler: handler,
	}

	for i := 0; i < size; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	slog.Info("worker pool started", "workers", size)
	return pool
}

// worker is the goroutine that processes jobs.
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			slog.Info("worker shutting down", "id", id)
			return
		case job, ok := <-p.jobs:
			if !ok {
				slog.Info("worker exiting (jobs channel closed)", "id", id)
				return
			}
			jobID := atomic.AddInt64(&p.jobCounter, 1)
			job.ID = jobID

			slog.Info("worker processing job", "worker_id", id, "job_id", jobID, "type", job.Type, "chat_id", job.ChatID)

			// Start progress reporting goroutine.
			progressCtx, progressCancel := context.WithCancel(p.ctx)
			go p.runProgress(progressCtx, job)

			// Execute the job with its own context.
			jobCtx, jobCancel := context.WithCancel(p.ctx)
			if job.Ctx != nil {
				jobCtx = job.Ctx
			}

			result := p.handler(jobCtx, job)

			jobCancel()
			progressCancel()

			// Send result back.
			select {
			case job.ResultChan <- result:
			case <-time.After(5 * time.Second):
				slog.Warn("timed out sending result", "job_id", jobID)
			}
		}
	}
}

// runProgress calls the progress callback every 10 seconds.
func (p *Pool) runProgress(ctx context.Context, job Job) {
	if job.ProgressFn == nil {
		return
	}
	start := time.Now()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Call immediately with 0s.
	job.ProgressFn(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := int(time.Since(start).Seconds())
			job.ProgressFn(elapsed)
		}
	}
}

// Enqueue submits a job to the pool and returns the job ID.
func (p *Pool) Enqueue(job Job) int64 {
	jobID := atomic.AddInt64(&p.jobCounter, 1)
	job.ID = jobID

	select {
	case p.jobs <- job:
		slog.Info("job enqueued", "job_id", jobID, "type", job.Type, "chat_id", job.ChatID)
	case <-p.ctx.Done():
		slog.Warn("pool is shut down, cannot enqueue job", "job_id", jobID)
		job.ResultChan <- Result{Error: fmt.Errorf("pool is shut down")}
	}

	return jobID
}

// Shutdown gracefully stops all workers.
func (p *Pool) Shutdown() {
	slog.Info("shutting down worker pool...")
	p.cancel()
	close(p.jobs)
	p.wg.Wait()
	slog.Info("worker pool shut down complete")
}
