package kafka_cg_handlers

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"
	config "github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type UpdateBannerHandler interface {
	Handle(ctx context.Context, msg handlers.Message) error
}

type BannerUpdateCGHandler struct {
	cfg     config.KafkaCGHandlerConfig
	logger  logger.Logger
	handler UpdateBannerHandler
}

func NewBannerUpdateCGHandler(logger logger.Logger, handler UpdateBannerHandler, cfg config.KafkaCGHandlerConfig) *BannerUpdateCGHandler {
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
	defer t.Stop()

	msgs := make([]*sarama.ConsumerMessage, 0, h.cfg.EventsBatchSize)

	for {
		ctx, cancel := context.WithTimeout(session.Context(), h.cfg.BatchMaxWait)

		for range h.cfg.EventsBatchSize {
			select {
			// if session context done, exiting immediatly
			case <-session.Context().Done():
				cancel()
				return nil
			// autocommit
			case <-t.C:
				cancel()
				session.Commit()
				continue
			// one batch max wait time exceeded
			case <-ctx.Done():
			// new message
			case message := <-claim.Messages():
				msgs = append(msgs, message)
			}
		}

		cancel()

		wg := sync.WaitGroup{}
		for _, msg := range msgs {

			wg.Add(1)
			go func(m *sarama.ConsumerMessage) {
				defer wg.Done()
				err := h.handler.Handle(session.Context(), handlers.Message{
					Key:   m.Key,
					Value: m.Value,
				},
				)
				if err != nil {
					h.logger.Error("can't proceed message", "err", err)
				}
				session.MarkMessage(m, "")
			}(msg)

		}
		// clean slice of messages after usage for a new batch
		msgs = msgs[:0]
		wg.Wait()
	}
}
