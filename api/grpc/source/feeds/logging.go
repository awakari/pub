package feeds

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"github.com/awakari/pub/util"
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

func (sl serviceLogging) Create(ctx context.Context, feed *Feed) (msg string, err error) {
	msg, err = sl.svc.Create(ctx, feed)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("feeds.Create(%+v): msg=%s, err=%s", feed, msg, err))
	return
}

func (sl serviceLogging) Read(ctx context.Context, url string) (feed *Feed, err error) {
	feed, err = sl.svc.Read(ctx, url)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("feeds.Read(%s): err=%s", url, err))
	return
}

func (sl serviceLogging) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	err = sl.svc.Delete(ctx, url, groupId, userId)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("feeds.Delete(%s, %s, %s): err=%s", url, groupId, userId, err))
	return
}

func (sl serviceLogging) ListUrls(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = sl.svc.ListUrls(ctx, filter, limit, cursor, order)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("feeds.ListUrls(%+v, %d, %s, %s): %d, err=%s", filter, limit, cursor, order, len(page), err))
	return
}
