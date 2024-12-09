package telegram

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

func (sl serviceLogging) Create(ctx context.Context, ch *Channel) (err error) {
	err = sl.svc.Create(ctx, ch)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.source.telegram.Create(%+v): ok", ch))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.source.telegram.Create(%+v): %s", ch, err))
	}
	return
}

func (sl serviceLogging) Read(ctx context.Context, link string) (ch *Channel, err error) {
	ch, err = sl.svc.Read(ctx, link)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.source.telegram.Read(%s): %+v", link, ch))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.source.telegram.Read(%s): %s", link, err))
	}
	return
}

func (sl serviceLogging) Delete(ctx context.Context, link string) (err error) {
	err = sl.svc.Delete(ctx, link)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.source.telegram.Delete(%s): ok", link))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.source.telegram.Delete(%s): %s", link, err))
	}
	return
}

func (sl serviceLogging) List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = sl.svc.List(ctx, filter, limit, cursor, order)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.source.telegram.List(%+v, %d, %s, %s): %d", filter, limit, cursor, order, len(page)))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.source.telegram.List(%+v, %d, %s, %s): %s", filter, limit, cursor, order, err))
	}
	return
}

func (sl serviceLogging) Login(ctx context.Context, code string, replicaIdx uint32) (success bool, err error) {
	success, err = sl.svc.Login(ctx, code, replicaIdx)
	switch err {
	case nil:
		sl.log.Debug(fmt.Sprintf("pub.grpc.source.telegram.Login(%s. %d): %t", code, replicaIdx, success))
	default:
		sl.log.Error(fmt.Sprintf("pub.grpc.source.telegram.Login(%s, %d): %s", code, replicaIdx, err))
	}
	return
}
