package events

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

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

func TestService_Publish(t *testing.T) {
	svc := NewService(NewClientMock())
	svc = NewLoggingMiddleware(svc, slog.Default())
	cases := map[string]struct {
		topic    string
		ackCount uint32
		err      error
	}{
		"ok": {
			topic:    "ok",
			ackCount: 42,
		},
		"fail": {
			topic: "fail",
			err:   ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ackCount, err := svc.Publish(context.TODO(), c.topic, []*pb.CloudEvent{})
			assert.Equal(t, c.ackCount, ackCount)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
