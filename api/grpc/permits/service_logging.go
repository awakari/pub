package permits

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

func (sl serviceLogging) GetUsage(ctx context.Context, groupId, userId string, subj model.Subject, out *model.Usage) (err error) {
	err = sl.svc.GetUsage(ctx, groupId, userId, subj, out)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("permits.GetUsage(%s, %s, %s): out=%+v, err=%s", groupId, userId, subj, out, err))
	return
}

func (sl serviceLogging) Request(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (p model.Permit, err error) {
	p, err = sl.svc.Request(ctx, groupId, userId, subj, count)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("permits.Request(%s, %s, %s, %d): %+v, err=%s", groupId, userId, subj, count, p, err))
	return
}

func (sl serviceLogging) Release(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (err error) {
	err = sl.svc.Release(ctx, groupId, userId, subj, count)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("permits.Release(%s, %s, %s, %d): err=%s", groupId, userId, subj, count, err))
	return
}
