package config

import "github.com/IBM/sarama"

type KafkaConsumerConfig struct {
	Addrs     []string
	SaramaCfg *sarama.Config
}
