package service

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) Preprocess(ctx context.Context, evt *pb.CloudEvent, groupId, userId string, internal bool) error {
	err := l.svc.Preprocess(ctx, evt, groupId, userId, internal)
	if err != nil {
		l.log.Error(fmt.Sprintf("service.Preprocess(%s, %s/%s, %t): %s", evt.Id, groupId, userId, internal, err))
		return err
	}
	return nil
}
