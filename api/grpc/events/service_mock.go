package events

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) SetStream(ctx context.Context, topic string, limit uint32) (err error) {
	switch topic {
	case "":
		err = ErrInvalid
	case "fail":
		err = ErrInternal
	}
	return
}

func (sm serviceMock) Publish(ctx context.Context, topic string, evts []*pb.CloudEvent) (ackCount uint32, err error) {

	return
}
