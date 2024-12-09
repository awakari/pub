package events

import (
	"context"
	"github.com/awakari/pub/api/grpc/ce"
	"github.com/awakari/pub/model"
)

type publisherMock struct{}

func NewPublisherMock() model.MessagesWriter {
	return publisherMock{}
}

func (mw publisherMock) Close() error {
	return nil
}

func (mw publisherMock) Write(ctx context.Context, msgs []*ce.CloudEvent) (ackCount uint32, err error) {
	for _, msg := range msgs {
		switch msg.Id {
		case "queue_fail":
			err = ErrInternal
		default:
			ackCount++
		}
		if err != nil {
			break
		}
	}
	return
}
