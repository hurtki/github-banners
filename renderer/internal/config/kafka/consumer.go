package kafka_config

import (
	"github.com/IBM/sarama"
)

type KafkaConsumerConfig struct {
	Addrs     []string
	SaramaCfg *sarama.Config
}

func NewKafkaConsumerConfig() KafkaConsumerConfig {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false
	return KafkaConsumerConfig{
		Addrs:     []string{"kafka:9092"},
		SaramaCfg: saramaCfg,
	}
}
