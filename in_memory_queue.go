package gogi

import (
	"context"
	"fmt"

	"sync"
	"time"
)

type InMemoryQueue struct {
	mu    sync.Mutex
	queue []Job
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{queue: make([]Job, 0)}
}

func (q *InMemoryQueue) SendJob(ctx context.Context, job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, job)
	return nil
}

func (q *InMemoryQueue) ReceiveJobs(ctx context.Context, handler func(Job) error) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				q.mu.Lock()
				if len(q.queue) == 0 {
					q.mu.Unlock()
					time.Sleep(100 * time.Millisecond)
					continue
				}
				job := q.queue[0]
				q.queue = q.queue[1:]
				q.mu.Unlock()

				if err := handler(job); err != nil {
					GetLogger().Debug(fmt.Sprintf("InMemoryQueue job error: %v", err))
				}
			}
		}
	}()
	return nil
}
