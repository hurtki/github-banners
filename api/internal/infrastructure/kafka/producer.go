package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type BannerProducer struct {
	producer sarama.SyncProducer
	topic    string
	logger   logger.Logger
}

func NewBannerProducer(brokers []string, topic string, cfg *sarama.Config, logger logger.Logger) (*BannerProducer, error) {
	fn := "internal.infrastructure.kafka.NewBannerProducer"
	var producer sarama.SyncProducer
	var err error
	for try := range 10 {
		producer, err = sarama.NewSyncProducer(brokers, cfg)
		if err != nil {
			logger.Warn("can't connect to kafka", "try", try+1, "err", err, "source", fn)
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("kafka producer init failed: %w", err)
	}

	return &BannerProducer{
		producer: producer,
		topic:    topic,
		logger:   logger.With("service", "kafka-infrastrcture"),
	}, nil
}

func (p *BannerProducer) Publish(ctx context.Context, info domain.LTBannerInfo) error {
	fn := "internal.infrastrcture.kafka.BannerProducer.Publish"
	event := GithubBannerInfoEvent{
		EventType:    "github_banner_info_ready",
		EventVersion: 1,
		ProducedAt:   time.Now().UTC(),
		Payload:      FromDomainBannerInfoToPayload(info),
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("unexpected error, when marshaling event", "source", fn, "err", err)
		return domain.ErrUnavailable
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(info.Username),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("can't send new event to kafka", "source", fn, "err", err)
		return domain.ErrUnavailable
	}

	return nil
}
