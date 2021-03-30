package rabbitmq

import (
	"errors"
	"time"
)

var (
	// ErrDisconnected to notify an error when client disconnected with rabbitmq server
	ErrDisconnected = errors.New("disconnected from rabbitmq, trying to reconnect")
	// ErrInvalidQueueNameProvided to notify an error if client provide a queue name which is not exists
	ErrInvalidQueueNameProvided = errors.New("invalid queue name provided")
	//
	ErrPushMessageFailed = errors.New("failed to push message to server")
)

const (
	ReconnectedDelay = 5 * time.Second
	ResendMsgDelay   = 5 * time.Second
)

const (
	ContentTypeTextPlain = "text/plain"
)
