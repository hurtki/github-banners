package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	config "github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

const (
	kafkaConnectionTries            = 10
	kafkaConnectionTimeBetweenTries = time.Second
)

type KafkaConsumerGroup struct {
	ctx    context.Context
	cg     sarama.ConsumerGroup
	logger logger.Logger
}

func NewKafkaConsumerGroup(ctx context.Context, logger logger.Logger, cfg config.KafkaConsumerConfig) (*KafkaConsumerGroup, error) {
	fn := "internal.infrastructure.kafka.NewKafkaConsumerGroup"

	var cg sarama.ConsumerGroup
	var err error
	for i := range kafkaConnectionTries {
		cg, err = sarama.NewConsumerGroup(cfg.Addrs, "banner-update-cg", cfg.SaramaCfg)
		if err != nil {
			logger.Warn("can't initialize consumer group", "try", i+1, "source", fn, "addrs", cfg.Addrs)
			if i == (kafkaConnectionTries - 1) {
				break
			}
			time.Sleep(kafkaConnectionTimeBetweenTries)
			continue
		} else {
			break
		}

	}
	if err != nil {
		return nil, fmt.Errorf("kafka consumer group init failed: %w", err)
	}

	return &KafkaConsumerGroup{
		ctx:    ctx,
		cg:     cg,
		logger: logger,
	}, nil
}

func (c *KafkaConsumerGroup) RegisterCGHandler(topics []string, handler sarama.ConsumerGroupHandler) error {
	err := c.cg.Consume(c.ctx, topics, handler)

	if err != nil {
		return fmt.Errorf("error occured when registering consumer group handler, %w", err)
	}

	return nil
}
