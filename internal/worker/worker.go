package worker

import (
	"context"
	"log"
	"math"
	"os/exec"
	"queueCtl/internal/config"
	"queueCtl/internal/model"
	"queueCtl/internal/storage"
	"sync"
	"time"
)

type Worker struct {
	ID     int
	Store  *storage.Store
	Config *config.Config
}

func New(id int, store *storage.Store, cfg *config.Config) *Worker {
	return &Worker{
		ID:     id,
		Store:  store,
		Config: cfg,
	}
}

// Run is the main loop for the worker.
// It polls for jobs and executes them.
func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Worker %d: Starting", w.ID)

	// Poll for jobs every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Context was canceled (shutdown signal)
			log.Printf("Worker %d: Shutting down...", w.ID)
			return
		case <-ticker.C: // Time to check for a job
			w.processJob()
		}
	}
}

// processJob finds and executes a single job.
func (w *Worker) processJob() {
	// Step 1: Find and lock a job
	job, err := w.Store.FindAndLock()
	if err != nil {
		log.Printf("Worker %d: Error finding job: %v", w.ID, err)
		return
	}
	if job == nil {
		return // No job found, just loop again
	}

	log.Printf("Worker %d: Processing job %s (command: %s)", w.ID, job.ID, job.Command)

	// Step 2: Execute the job's command
	// We use "sh -c" to allow for complex commands
	cmd := exec.Command("sh", "-c", job.Command)
	execErr := cmd.Run() // This blocks until the command finishes

	// Step 3: Update the job based on the result
	job.UpdatedAt = time.Now()

	if execErr == nil {
		// --- SUCCESS ---
		job.State = model.StateCompleted
		log.Printf("Worker %d: Job %s completed successfully", w.ID, job.ID)
	} else {
		// --- FAILURE ---
		log.Printf("Worker %d: Job %s failed: %v", w.ID, job.ID, execErr)
		
		if job.Attempts >= job.MaxRetries {
			// --- DEAD (Max retries reached) ---
			job.State = model.StateDead
			log.Printf("Worker %d: Job %s moved to Dead Letter Queue (DLQ)", w.ID, job.ID)
		} else {
			// --- FAILED (Retryable) ---
			job.State = model.StateFailed
			
			// Calculate exponential backoff
			delay := math.Pow(w.Config.BackoffBase, float64(job.Attempts))
			job.NextRunAt = time.Now().Add(time.Second * time.Duration(delay))
			
			log.Printf("Worker %d: Job %s will retry in %.0fs", w.ID, job.ID, delay)
		}
	}

	// Step 4: Save the job's final state
	if err := w.Store.UpdateJob(job); err != nil {
		log.Printf("Worker %d: Error updating job %s: %v", w.ID, job.ID, err)
	}
}