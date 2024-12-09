package events

import (
	"context"
	"github.com/awakari/pub/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"io"
)

type publisher struct {
	stream Service_PublishClient
	queue  string
}

func newPublisher(stream Service_PublishClient, queue string) model.MessagesWriter {
	return publisher{
		stream: stream,
		queue:  queue,
	}
}

func (mw publisher) Close() (err error) {
	err = mw.stream.CloseSend()
	if err != nil {
		err = decodeError(err)
	}
	return
}

func (mw publisher) Write(ctx context.Context, msgs []*pb.CloudEvent) (ackCount uint32, err error) {
	req := PublishRequest{
		Topic: mw.queue,
		Evts:  msgs,
	}
	err = mw.stream.Send(&req)
	var resp *PublishResponse
	if err == nil || err == io.EOF {
		resp, err = mw.stream.Recv()
	}
	if err != nil {
		err = decodeError(err)
	}
	if resp != nil {
		ackCount = resp.AckCount
	}
	return
}
