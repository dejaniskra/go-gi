package gogi

import (
	"context"
	"sync"
)

type InMemoryPubSub struct {
	subscribers map[string][]func(*Event) error
	mu          sync.RWMutex
}

func NewInMemoryPubSub() *InMemoryPubSub {
	return &InMemoryPubSub{
		subscribers: make(map[string][]func(*Event) error),
	}
}

func (ps *InMemoryPubSub) Publish(ctx context.Context, event *Event) error {
	ps.mu.RLock()
	handlers := ps.subscribers[event.Topic]
	ps.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event) // fire-and-forget
	}
	return nil
}

func (ps *InMemoryPubSub) Subscribe(ctx context.Context, topic string, handler func(*Event) error) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subscribers[topic] = append(ps.subscribers[topic], handler)
	return nil
}

func (ps *InMemoryPubSub) Close() error {
	ps.subscribers = nil
	return nil
}
