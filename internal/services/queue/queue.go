package queue

import (
	"context"
)

// MessageQueue defines the contract for our async messaging
type MessageQueue interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Subscribe(topic string) (<-chan interface{}, error)
}
