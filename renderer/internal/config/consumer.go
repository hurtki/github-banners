package config

import (
	"strings"

	"github.com/IBM/sarama"
)

type KafkaConsumerConfig struct {
	Addrs     []string
	SaramaCfg *sarama.Config
}

func NewKafkaConsumerConfig() KafkaConsumerConfig {
	saramaCfg := sarama.NewConfig()

	// turning off autocommit
	// we implement it ourselfs, because in some solutions it will be harmfull
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false
	saramaCfg.Version = sarama.V4_1_0_0

	addrs := []string{}
	addrsVariable := getEnv("KAFKA_BROKERS_ADDRS", "kafka:9092")

	addrs = append(addrs, strings.Split(addrsVariable, ",")...)

	return KafkaConsumerConfig{
		Addrs:     addrs,
		SaramaCfg: saramaCfg,
	}
}
