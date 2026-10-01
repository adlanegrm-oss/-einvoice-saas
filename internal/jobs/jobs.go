package jobs

import (
	"sync"
	"time"
)

type JobStatus string

const (
	StatusQueued     JobStatus = "QUEUED"
	StatusProcessing JobStatus = "PROCESSING"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusFailed     JobStatus = "FAILED"
)

type Job struct {
	ID        string    `json:"job_id"`
	Type      string    `json:"type"` // CLEARANCE_SUBMISSION, AI_OCR
	InvoiceID string    `json:"invoice_id"`
	Status    JobStatus `json:"status"`
	Progress  int       `json:"progress"` // 0 - 100%
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tracker struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewTracker() *Tracker {
	return &Tracker{
		jobs: make(map[string]*Job),
	}
}

func (t *Tracker) TrackJob(job *Job) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.jobs[job.ID] = job
}

func (t *Tracker) GetActiveJobs() []*Job {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var active []*Job
	for _, j := range t.jobs {
		if j.Status == StatusQueued || j.Status == StatusProcessing {
			active = append(active, j)
		}
	}
	return active
}
