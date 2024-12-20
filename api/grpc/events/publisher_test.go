package events

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestMessagesWriter_Close(t *testing.T) {
	mw := newPublisher(newPublishStreamMock(), "queue0")
	err := mw.Close()
	assert.Nil(t, err)
}

func TestMessagesWriter_Write(t *testing.T) {
	cases := map[string]struct {
		queue    string
		msgs     []*pb.CloudEvent
		ackCount uint32
		err      error
	}{
		"1 => ack 1": {
			queue: "queue0",
			msgs: []*pb.CloudEvent{
				{
					Id: "msg0",
				},
			},
			ackCount: 1,
		},
		"3 => ack 2": {
			queue: "queue0",
			msgs: []*pb.CloudEvent{
				{
					Id: "msg0",
				},
				{
					Id: "msg1",
				},
				{
					Id: "msg2",
				},
			},
			ackCount: 2,
		},
		"send eof": {
			queue: "send_eof",
			msgs: []*pb.CloudEvent{
				{
					Id: "msg0",
				},
			},
			err: io.EOF,
		},
		"recv fail": {
			queue: "recv_fail",
			msgs: []*pb.CloudEvent{
				{
					Id: "msg0",
				},
			},
			err: ErrInternal,
		},
		"recv eof": {
			queue: "recv_eof",
			msgs: []*pb.CloudEvent{
				{
					Id: "msg0",
				},
			},
			err: io.EOF,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			mw := newPublisher(newPublishStreamMock(), c.queue)
			ackCount, err := mw.Write(context.TODO(), c.msgs)
			assert.Equal(t, c.ackCount, ackCount)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
