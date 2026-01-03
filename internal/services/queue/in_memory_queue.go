package queue

import (
	"errors"
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

func (q *InMemoryQueue) Publish(topic string, message interface{}) error {
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

	// Non-blocking send (optional: drop message if full, or block)
	select {
	case ch <- message:
		return nil
	default:
		return errors.New("topic queue is full")
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
