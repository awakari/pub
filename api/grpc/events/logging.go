package events

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"log/slog"
)

type loggingMiddleware struct {
	svc Service
	log *slog.Logger
}

func NewLoggingMiddleware(svc Service, log *slog.Logger) Service {
	return loggingMiddleware{
		svc: svc,
		log: log,
	}
}

func (lm loggingMiddleware) SetStream(ctx context.Context, topic string, limit uint32) (err error) {
	err = lm.svc.SetStream(ctx, topic, limit)
	lm.log.Debug(fmt.Sprintf("events.SetStream(topic=%s, limit=%d): err=%s", topic, limit, err))
	return
}

func (lm loggingMiddleware) NewPublisher(ctx context.Context, topic string) (p model.MessagesWriter, err error) {
	p, err = lm.svc.NewPublisher(ctx, topic)
	lm.log.Debug(fmt.Sprintf("events.Publish(topic=%s): err=%s", topic, err))
	return
}
