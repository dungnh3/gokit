package pubsub

import "context"

// Publisher is a generic interface to wrapper how we want to publish message.
type Publisher interface {
	Publish(ctx context.Context, key string, value []byte) error
}

// Subscriber is generate interface to wrapper how we want to consume message.
type Subscriber interface {
	// Start will return a channel of raw message
	Start() <-chan SubscriberMessage
	// Error will contain any error returned from the consumer connection
	Error() error
	// Stop will initiate a graceful shutdown of the subcriber connection
	Stop() error
}

// SubscriberMessage is a struct to encapsulate subscriber messages and provide
// a mechanism for acknowledging messages _after_ they've been processed.
type SubscriberMessage interface {
	Message() []byte
	Key() []byte
	Done() error
}
