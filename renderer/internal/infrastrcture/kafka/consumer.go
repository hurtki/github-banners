package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type KafkaConsumer struct {
	cg                       sarama.ConsumerGroup
	logger                   logger.Logger
	cg_banner_update_handler sarama.ConsumerGroupHandler
}

func NewKafkaConsumer(logger logger.Logger, cfg config.KafkaConsumerConfig, cg_banner_update_handler sarama.ConsumerGroupHandler) (*KafkaConsumer, error) {
	fn := "internal.infrastructure.kafka.NewKafkaConsumer"
	var cg sarama.ConsumerGroup
	var err error
	for i := range 10 {
		cg, err = sarama.NewConsumerGroup(cfg.Addrs, "banner-update", cfg.SaramaCfg)
		if err != nil {
			logger.Warn("can't initialize consumer group", "try", i+1, "source", fn)
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("kafka consumer group init failed: %w", err)
	}
	return &KafkaConsumer{
		cg:                       cg,
		logger:                   logger,
		cg_banner_update_handler: cg_banner_update_handler,
	}, nil
}

func (c *KafkaConsumer) StartListening(ctx context.Context) {
	err := c.cg.Consume(ctx, []string{"banner-update"}, c.cg_banner_update_handler)
	if err != nil {
		fmt.Println("error occured when starting consuming")
	}
}
