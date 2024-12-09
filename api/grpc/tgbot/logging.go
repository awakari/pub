package tgbot

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"log/slog"
)

type serviceLogging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return serviceLogging{
		svc: svc,
		log: log,
	}
}

func (sl serviceLogging) ReadChannel(ctx context.Context, link string) (ch *Channel, err error) {
	ch, err = sl.svc.ReadChannel(ctx, link)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.tgbot.ReadChannel(%s): %+v", link, ch))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.tgbot.ReadChannel(%s): %s", link, err))
	}
	return
}

func (sl serviceLogging) ListChannels(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = sl.svc.ListChannels(ctx, filter, limit, cursor, order)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.tgbot.ListChannels(%+v, %d, %s, %s): %d", filter, limit, cursor, order, len(page)))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.tgbot.ListChannels(%+v, %d, %s, %s): %s", filter, limit, cursor, order, err))
	}
	return
}
