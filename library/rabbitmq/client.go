package rabbitmq

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"os"
	"runtime"
	"sync"
	"time"
)

type Logger interface {
	Info(message string, opts ...interface{})
	Warn(message string, opts ...interface{})
	Error(message string, opts ...interface{})
}

type Client struct {
	// logger is used to log any information, maybe info, warning, error or something like that.
	// logger is no mandatory and could be replace with any other structured logger like Zap/Logrus.
	logger *zap.Logger

	// Connection manages the serialization and deserialization of frames from IO
	// and dispatches the frames to the appropriate channel.  All RPC methods and
	// asynchronous Publishing, Delivery, Ack, Nack and Return messages are
	// multiplexed on this channel.  There must always be active receivers for
	// every asynchronous message on this connection.
	connection *amqp.Connection

	// Channel represents an AMQP channel. Used as a context for valid message
	// exchange.  Errors on methods with this Channel as a receiver means this channel
	// should be discarded and a new channel established.
	channel *amqp.Channel

	// done channel is triggered once the server has received the shutdown signal, which will stop
	// the client from trying to reconnect to RabbitMQ server.
	done chan os.Signal

	// notifyClose tell the reconnect method that the connections closed and the client needs to reconnect.
	notifyClose chan *amqp.Error

	// notifyConfirm us used to acknowledge that the data was pushed to the server.
	notifyConfirm chan amqp.Confirmation

	isConnected bool
	alive       bool
	threads     int

	// wg is used to gracefull shutdown the server once all messages are processed.
	wg *sync.WaitGroup
}

func New(address string, logger *zap.Logger, done chan os.Signal) (*Client, error) {
	threads := runtime.GOMAXPROCS(0)
	if numCPU := runtime.NumCPU(); numCPU > threads {
		threads = numCPU
	}

	client := &Client{
		logger:      logger,
		done:        done,
		isConnected: false,
		alive:       true,
		threads:     threads,
		wg:          &sync.WaitGroup{},
	}
	client.wg.Add(threads)

	return client, nil
}

func (c *Client) handleReconnect(listenQueue string, addr string) {
	for c.alive {
		c.isConnected = true
		c.logger.Info("Trying to connect to RabbitMQ..",
			zap.Any("address", addr),
		)

		retryCount := 0
		for !c.connect(listenQueue, addr) {
			if c.alive {
				return
			}

			select {
			case <-c.done:
				return
			case <-time.After(ReconnectedDelay + time.Duration(retryCount)*time.Second):
				c.logger.Warn("disconnected from RabbitMQ and failed to connect")
				retryCount++
			}

			select {
			case <-c.done:
				return
			case <-c.notifyClose:

			}
		}
	}
}

func (c *Client) connect(listenQueue string, addr string) bool {
	conn, err := amqp.Dial(addr)
	if err != nil {
		c.logger.Error("Failed to connect with RabbitMQ server",
			zap.Any("address", addr),
			zap.Any("error", err),
		)
		return false
	}

	channel, err := conn.Channel()
	if err != nil {
		c.logger.Error("Failed to connect with channel",
			zap.Any("error", err),
		)
		return false
	}

	c.changeConnection(conn, channel)
	c.isConnected = true
	return true
}

func (c *Client) changeConnection(connection *amqp.Connection, channel *amqp.Channel) {
	c.connection = connection
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.notifyConfirm = make(chan amqp.Confirmation)
	c.channel.NotifyClose(c.notifyClose)
	c.channel.NotifyPublish(c.notifyConfirm)
}

func (c *Client) Push(message []byte) error {
	if !c.isConnected {
		return ErrPushMessageFailed
	}

	for {
		if err := c.PushWithUnsafe(message); err != nil {
			if err == ErrDisconnected {
				continue
			}
			return err
		}

		select {
		case confirm := <-c.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <-time.After(ResendMsgDelay):
			c.logger.Info("resend message..", zap.Any("after", ResendMsgDelay))
		}
	}
}

// PushWithUnsafe will push to the queue without checking for confirmation. It
// return an error if it fails to connect. No guarantees are provided for whether
// the server will receive the message.
func (c *Client) PushWithUnsafe(message []byte) error {
	if !c.isConnected {
		return ErrDisconnected
	}

	return c.channel.Publish(
		/* exchange */ "",
		/* rounting key */ "",
		false,
		false,
		amqp.Publishing{
			Headers:         nil,
			ContentType:     ContentTypeTextPlain,
			ContentEncoding: "",
			DeliveryMode:    0,
			Priority:        0,
			CorrelationId:   "",
			ReplyTo:         "",
			Expiration:      "",
			MessageId:       "",
			Timestamp:       time.Time{},
			Type:            "",
			UserId:          "",
			AppId:           "",
			Body:            message,
		},
	)
}
