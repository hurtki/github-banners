package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type UpdateBannerHandler interface {
	Handle(ctx context.Context, msg handlers.Message) error
}

type ConsumerGroupHandler struct {
	logger  logger.Logger
	handler UpdateBannerHandler
}

func NewConsumeGroupHandler(logger logger.Logger, handler UpdateBannerHandler) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		logger:  logger,
		handler: handler,
	}
}

func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	session.Claims()
	h.logger.Info("new active session")
	return nil
}
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("session ended")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	t := time.NewTicker(time.Second * 30)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		msgs := make([]*sarama.ConsumerMessage, 0)
		for range 10 {
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
