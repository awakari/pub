package sites

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

func (sl serviceLogging) Create(ctx context.Context, site *Site) (err error) {
	err = sl.svc.Create(ctx, site)
	sl.log.Log(ctx, logLevel(err), fmt.Sprintf("sites.Create(%+v): err=%s", site, err))
	return
}

func (sl serviceLogging) Read(ctx context.Context, addr string) (site *Site, err error) {
	site, err = sl.svc.Read(ctx, addr)
	sl.log.Log(ctx, logLevel(err), fmt.Sprintf("sites.Read(%s): site=%+v, err=%s", addr, site, err))
	return
}

func (sl serviceLogging) Delete(ctx context.Context, addr, groupId, userId string) (err error) {
	err = sl.svc.Delete(ctx, addr, groupId, userId)
	sl.log.Log(ctx, logLevel(err), fmt.Sprintf("sites.Delete(%s): err=%s", addr, err))
	return
}

func (sl serviceLogging) List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = sl.svc.List(ctx, filter, limit, cursor, order)
	sl.log.Log(ctx, logLevel(err), fmt.Sprintf("sites.List(filter=%+v, limit=%d, cursor=%s, order=%s): %d, err=%s", filter, limit, cursor, order, len(page), err))
	return
}

func logLevel(err error) (ll slog.Level) {
	switch err {
	case nil:
		ll = slog.LevelDebug
	default:
		ll = slog.LevelError
	}
	return
}
