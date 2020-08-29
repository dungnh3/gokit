package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
)

var (
	RequiredAcks = sarama.WaitForAll

	ErrTopicIsRequired = errors.New("topic name is required")
)

// Config is the configured files may be in kafka producer or consumer
type Config struct {
	// BrokerHost ..
	BrokerHost struct {
		// Cluster is a set of server kafka
		Cluster []string
		// ClusterRetry is a set of server kafka that is used to retry push message if error occur
		ClusterRetry []string
		// ClusterDLQ is a set of server kafka that is used to hold all message if error occur while consume error
		// DLQ that 's mean dead letter queue
		ClusterDLQ []string
	}

	// Topic ..
	Topic []string

	// MaxRetry ..
	MaxRetry int

	// Partition ..
	Partition int32

	// Config ..
	Config *sarama.Config
}
