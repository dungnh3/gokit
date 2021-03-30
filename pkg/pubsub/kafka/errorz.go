package kafka

import "errors"

var (
	// ErrConsumerGroupIsInvalid return a new error with message consumer group is invalid
	ErrConsumerGroupIsInvalid = errors.New("consumer group is invalid")
	// ErrHandlerFuncIsInvalid return a new error with message handler func is invalid
	ErrHandlerFuncIsInvalid = errors.New("handler func is invalid")
	// ErrNumberOfWorkerHandlerInvalid
	ErrNumberOfWorkerHandlerInvalid = errors.New("number of worker handler must greater than 0")
)
