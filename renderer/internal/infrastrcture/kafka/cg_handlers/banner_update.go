package kafka_cg_handlers

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"
	kafka_config "github.com/hurtki/github-banners/renderer/internal/config/kafka"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type UpdateBannerHandler interface {
	Handle(ctx context.Context, msg handlers.Message) error
}

type BannerUpdateCGHandler struct {
	cfg     kafka_config.KafkaCGHandlerConfig
	logger  logger.Logger
	handler UpdateBannerHandler
}

func NewBannerUpdateCGHandler(logger logger.Logger, handler UpdateBannerHandler, cfg kafka_config.KafkaCGHandlerConfig) *BannerUpdateCGHandler {
	return &BannerUpdateCGHandler{
		logger:  logger.With("service", "banner-update-cg-handler"),
		handler: handler,
		cfg:     cfg,
	}
}

func (h *BannerUpdateCGHandler) Setup(sess sarama.ConsumerGroupSession) error {
	for topic, partitions := range sess.Claims() {
		h.logger.Info("New consumer group session", "topic", topic, "partitions", partitions)
	}
	return nil
}
func (h *BannerUpdateCGHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	return func() error { sess.Commit(); return nil }()
}

func (h *BannerUpdateCGHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	t := time.NewTicker(h.cfg.AutoCommitInterval)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), h.cfg.BatchMaxWait)
		msgs := make([]*sarama.ConsumerMessage, 0, h.cfg.EventsBatchSize)

		for range h.cfg.EventsBatchSize {
			select {
			case <-session.Context().Done():
				cancel()
				return nil
			case <-t.C:
				cancel()
				session.Commit()
				continue
			case <-ctx.Done():
			case message := <-claim.Messages():
				msgs = append(msgs, message)
			}
		}

		cancel()

		wg := sync.WaitGroup{}
		for _, msg := range msgs {
			wg.Go(func() {
				err := h.handler.Handle(session.Context(), handlers.Message{
					Key:   msg.Key,
					Value: msg.Value,
				},
				)
				if err != nil {
					h.logger.Error("can't proceed message", "err", err)
				}
				session.MarkMessage(msg, "")
			})
		}
		wg.Wait()
	}
}
