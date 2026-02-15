package kafka_config

import "time"

type KafkaCGHandlerConfig struct {
	AutoCommitInterval time.Duration
	EventsBatchSize    int
	BatchMaxWait       time.Duration
}

func NewKafkaCGHandlerConfig() KafkaCGHandlerConfig {
	return KafkaCGHandlerConfig{
		AutoCommitInterval: time.Second * 1,
		EventsBatchSize:    10,
		BatchMaxWait:       time.Second * 3,
	}
}
