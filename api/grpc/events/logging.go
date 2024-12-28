package events

import (
	"context"
	"fmt"
	"github.com/awakari/pub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
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

func (lm loggingMiddleware) Publish(ctx context.Context, topic string, evts []*pb.CloudEvent) (ackCount uint32, err error) {
	ackCount, err = lm.svc.Publish(ctx, topic, evts)
	ll := util.LogLevel(err)
	lm.log.Log(ctx, ll, fmt.Sprintf("events.Publish(%s, %d): ack=%d, err=%s", topic, len(evts), ackCount, err))
	return
}
