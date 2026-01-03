package queue

import (
	"context"
	"sync"
)

type InMemoryQueue struct {
	topics map[string]chan interface{}
	mu     sync.RWMutex
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		topics: make(map[string]chan interface{}),
	}
}

func (q *InMemoryQueue) Publish(ctx context.Context, topic string, message interface{}) error {
	q.mu.RLock()
	ch, exists := q.topics[topic]
	q.mu.RUnlock()
	if !exists {
		// If topic doesn't exist, create it on the fly (or return error depending on strictness)
		// For this MVP, let's create it safely.
		q.mu.Lock()
		ch = make(chan interface{}, 100) // buffered channel
		q.topics[topic] = ch
		q.mu.Unlock()
	}

	// Blocking send with Context Cancellation
	// In case the channel is full, we wait until there's space or context is done
	// In a production system, consider using a more robust queuing mechanism to handle backpressure and retries.
	select {
	case ch <- message:
		return nil
	case <-ctx.Done():
		// If the context (request) times out before we can push to queue, return error
		return ctx.Err()
	}
}

func (q *InMemoryQueue) Subscribe(topic string) (<-chan interface{}, error) {
	q.mu.RLock()
	ch, exists := q.topics[topic]
	q.mu.RUnlock()
	if !exists {
		// Create topic if it doesn't exist
		q.mu.Lock()
		ch = make(chan interface{}, 100) // buffered channel
		q.topics[topic] = ch
		q.mu.Unlock()
	}
	return ch, nil
}
