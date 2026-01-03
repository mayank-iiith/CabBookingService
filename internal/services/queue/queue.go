package queue

// MessageQueue defines the contract for our async messaging
type MessageQueue interface {
	Publish(topic string, message interface{}) error
	Subscribe(topic string) (<-chan interface{}, error)
}
