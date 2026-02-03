package config

import "github.com/IBM/sarama"

func NewProducerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0

	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Idempotent = true
	cfg.Producer.Return.Successes = true

	//required for Idempotent producer ordering
	cfg.Net.MaxOpenRequests = 1

	return cfg
}
