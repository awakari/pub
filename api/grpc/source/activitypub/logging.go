package activitypub

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"github.com/awakari/pub/util"
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

func (l logging) Create(ctx context.Context, addr, groupId, userId string) (url string, err error) {
	url, err = l.svc.Create(ctx, addr, groupId, userId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.Create(addr=%s, groupId=%s, userId=%s): %s, %s", addr, groupId, userId, url, err))
	return
}

func (l logging) Read(ctx context.Context, url string) (a *Source, err error) {
	a, err = l.svc.Read(ctx, url)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.Read(url=%s): %+v, %s", url, a, err))
	return
}

func (l logging) List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = l.svc.List(ctx, filter, limit, cursor, order)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub..List(filter=%+v, limit=%d, cursor=%s, order=%s): %d, %s", filter, limit, cursor, order, len(page), err))
	return
}

func (l logging) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	err = l.svc.Delete(ctx, url, groupId, userId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.Delete(url=%s, groupId=%s, userId=%s): %s", url, groupId, userId, err))
	return
}
