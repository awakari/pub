package tgbot

import (
	"context"
	"github.com/awakari/pub/model"
)

type svcMock struct {
}

func NewServiceMock() Service {
	return svcMock{}
}

func (sm svcMock) Authenticate(ctx context.Context, token []byte) (err error) {
	//TODO implement me
	panic("implement me")
}

func (sm svcMock) ReadChannel(ctx context.Context, link string) (ch *Channel, err error) {
	//TODO implement me
	panic("implement me")
}

func (sm svcMock) ListChannels(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	//TODO implement me
	panic("implement me")
}
