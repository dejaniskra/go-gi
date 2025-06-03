package gogi

import "context"

type Event struct {
	Topic   string
	Payload []byte
}

type PubSub interface {
	Publish(ctx context.Context, event *Event) error
	Subscribe(ctx context.Context, topic string, handler func(*Event) error) error
	Close() error
}
