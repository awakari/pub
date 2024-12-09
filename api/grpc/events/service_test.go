package events

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestService_NewPublisher(t *testing.T) {
	svc := NewService(NewClientMock())
	svc = NewLoggingMiddleware(svc, slog.Default())
	cases := map[string]error{
		"ok": nil,
	}
	for k, expectedErr := range cases {
		t.Run(k, func(t *testing.T) {
			qw, err := svc.NewPublisher(context.TODO(), k)
			assert.ErrorIs(t, err, expectedErr)
			if err == nil {
				assert.NotNil(t, qw)
			}
		})
	}
}

func TestService_SetStream(t *testing.T) {
	svc := NewService(NewClientMock())
	svc = NewLoggingMiddleware(svc, slog.Default())
	cases := map[string]struct {
		topic string
		limit uint32
		err   error
	}{
		"ok": {
			topic: "ok",
		},
		"empty": {
			err: ErrInvalid,
		},
		"fail": {
			topic: "fail",
			err:   ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.SetStream(context.TODO(), c.topic, c.limit)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
