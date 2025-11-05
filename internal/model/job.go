package model

import "time"

const (
    StatePending    = "pending"
    StateProcessing = "processing"
    StateCompleted  = "completed"
    StateFailed     = "failed"
    StateDead       = "dead"
)

type Job struct {
    ID          string    `json:"id"`
    Command     string    `json:"command"`
    State       string    `json:"state"`
    Attempts    int       `json:"attempts"`
    MaxRetries  int       `json:"max_retries"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    NextRunAt   time.Time `json:"-"` // We add this for backoff, ignore in JSON
}