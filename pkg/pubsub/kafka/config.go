package kafka

type Config struct {
	Brokers string `json:"brokers" mapstructure:"brokers" yaml:"brokers"`
}

type ConsumeWorker struct {
	Topics string
}
