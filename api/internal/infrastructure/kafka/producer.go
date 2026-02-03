package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/hurtki/github-banners/api/internal/domain"
)


type BannerProducer struct {
	producer sarama.SyncProducer
	topic	string
}

func NewBannerProducer(brokers []string, topic string, cfg *sarama.Config) (*BannerProducer, error) {
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka producer init failed: %w", err)
	}

	return &BannerProducer{
		producer: producer,
		topic: topic,
	}, nil
}

func (p *BannerProducer) Publish(ctx context.Context, info domain.GithubUserBannerInfo)	error {
	event := GithubBannerInfoEvent{
		EventType: "github_banner_info_ready",
		EventVersion: 1,
		ProducedAt: time.Now().UTC(),
		Payload: Payload {
			Username: info.Username,
			BannerType: domain.BannerTypesBackward[info.BannerType],
			StoragePath: info.StoragePath,
			Stats: info.Stats,
			FetchedAt: time.Now().UTC(),
		},
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal kafka event failed: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key: sarama.StringEncoder(info.Username),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send kafka message failed: %w", err)
	}

	return nil
}
