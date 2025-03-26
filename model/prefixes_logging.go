package model

import (
	"context"
	"fmt"
	"github.com/awakari/pub/util"
	"log/slog"
)

type prefixesLogging[T any] struct {
	prefixes Prefixes[T]
	log      *slog.Logger
}

func NewPrefixesLogging[T any](prefixes Prefixes[T], log *slog.Logger) Prefixes[T] {
	return prefixesLogging[T]{
		prefixes: prefixes,
		log:      log,
	}
}

func (pl prefixesLogging[T]) Put(ctx context.Context, prefix string, v T) (err error) {
	err = pl.prefixes.Put(ctx, prefix, v)
	pl.log.Log(context.TODO(), util.LogLevel(err), fmt.Sprintf("prefixes.Put(%s, %+v): %s", prefix, v, err))
	return
}

func (pl prefixesLogging[T]) FindOne(ctx context.Context, input string) (prefix string, v T, err error) {
	prefix, v, err = pl.prefixes.FindOne(ctx, input)
	pl.log.Log(context.TODO(), util.LogLevel(err), fmt.Sprintf("prefixes.FindOne(%s): %s, _, %s", input, prefix, err))
	return
}
