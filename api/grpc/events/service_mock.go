package events

import (
	"context"
	"github.com/awakari/pub/model"
)

type serviceMock struct {
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

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) NewPublisher(ctx context.Context, topic string) (mw model.MessagesWriter, err error) {
	switch topic {
	case "fail":
		err = ErrInternal
	default:
		mw = NewPublisherMock()
	}
	return
}
