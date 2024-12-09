package limits

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

func (sl serviceLogging) Get(ctx context.Context, groupId, userId string, subj model.Subject) (l model.Limit, err error) {
	l, err = sl.svc.Get(ctx, groupId, userId, subj)
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("limits.Get(%s, %s, %s): %+v, err=%s", groupId, userId, subj, l, err))
	return
}
