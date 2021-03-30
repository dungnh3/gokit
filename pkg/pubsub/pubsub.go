package pubsub

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, message []byte, opts ...interface{}) error
}

type Subscriber interface {
	// Start
	Start() error
	// Err will contain any errors returned from the consumer connection.
	Error() error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}

// SubscriberMessage is a struct to encapsulate subscriber messages and provide
// a mechanism for acknowledging messages _after_ they've been processed.
type SubscriberMessage interface {
	Message() []byte
	Key() []byte
	Done() error
}
