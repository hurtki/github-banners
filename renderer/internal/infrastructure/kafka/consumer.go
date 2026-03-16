package kafka

import (
	"context"
	"fmt"
	"sync"
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
	cg     sarama.ConsumerGroup
	logger logger.Logger

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func NewKafkaConsumerGroup(logger logger.Logger, cfg config.KafkaConsumerConfig) (*KafkaConsumerGroup, error) {
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

	logger.Info("initialized consumer group successfully", "addrs", cfg.Addrs, "source", fn)

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumerGroup{
		cg:     cg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
	}, nil
}

func (c *KafkaConsumerGroup) RegisterCGHandler(topics []string, handler sarama.ConsumerGroupHandler) {
	c.wg.Go(func() {
		c.registerCGHandler(topics, handler)
	})
}

func (c *KafkaConsumerGroup) registerCGHandler(topics []string, handler sarama.ConsumerGroupHandler) {
	err := c.cg.Consume(c.ctx, topics, handler)
	if err != nil {
		c.logger.Error("error occurred, when registering consumer group handler", "err", err, "topics", topics)
	}
}

func (c *KafkaConsumerGroup) Close(ctx context.Context) error {
	c.cancel()
	done := make(chan error)
	go func() {
		c.wg.Wait()
		done <- c.cg.Close()
	}()
	select {
	case <-ctx.Done():
		c.logger.Warn("couldn't shutdown in time, exiting", "ctxErr", ctx.Err())
		return ctx.Err()
	case err := <-done:
		if err == nil {
			c.logger.Info("successfully shutted down")
		}
		return err
	}
}
