package gogi

import (
	"context"
)

// Job represents a unit of work.
type Job struct {
	ID   string                 // Optional ID (can be generated)
	Type string                 // e.g., "email", "image_processing"
	Data map[string]interface{} // Arbitrary payload
}

// JobQueue defines the interface all queue backends must implement.
type JobQueue interface {
	SendJob(ctx context.Context, job Job) error
	ReceiveJobs(ctx context.Context, handler func(Job) error) error
}
