package events

import (
	"context"
	"github.com/awakari/pub/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type publisherMock struct{}

func NewPublisherMock() model.MessagesWriter {
	return publisherMock{}
}

func (mw publisherMock) Close() error {
	return nil
}

func (mw publisherMock) Write(ctx context.Context, msgs []*pb.CloudEvent) (ackCount uint32, err error) {
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
