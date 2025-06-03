package gogi

import (
	"context"

	"sync"
)

type InMemoryJobQueue struct {
	queue []*Job
	mu    sync.Mutex
}

func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{
		queue: make([]*Job, 0),
	}
}

func (q *InMemoryJobQueue) SendJob(ctx context.Context, job *Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, job)
	return nil
}

func (q *InMemoryJobQueue) ReceiveJobs(ctx context.Context, handler func(*Job) error) error {
	for {
		q.mu.Lock()
		if len(q.queue) == 0 {
			q.mu.Unlock()
			break
		}
		job := q.queue[0]
		q.queue = q.queue[1:]
		q.mu.Unlock()
		if err := handler(job); err != nil {
			return err
		}
	}
	return nil
}
