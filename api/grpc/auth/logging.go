package auth

import (
	"context"
	"fmt"
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

func (l logging) Authenticate(ctx context.Context, userId, token string) (err error) {
	err = l.svc.Authenticate(ctx, userId, token)
	l.log.Debug(fmt.Sprintf("Authenticate(%s, _): %s", userId, err))
	return
}
