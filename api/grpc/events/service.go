package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type Service interface {
	SetStream(ctx context.Context, topic string, limit uint32) (err error)
	Publish(ctx context.Context, topic string, evts []*pb.CloudEvent) (ackCount uint32, err error)
}

type service struct {
	client ServiceClient
}

// ErrInternal indicates some unexpected internal failure.
var ErrInternal = errors.New("events: internal failure")

var ErrInvalid = errors.New("events: invalid request")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) SetStream(ctx context.Context, topic string, limit uint32) (err error) {
	_, err = svc.client.SetStream(ctx, &SetStreamRequest{
		Topic: topic,
		Limit: limit,
	})
	err = decodeError(err)
	return
}

func (svc service) Publish(ctx context.Context, topic string, evts []*pb.CloudEvent) (ackCount uint32, err error) {
	var resp *PublishResponse
	resp, err = svc.client.PublishBatch(ctx, &PublishRequest{
		Topic: topic,
		Evts:  evts,
	})
	if resp != nil {
		ackCount = resp.AckCount
	}
	err = decodeError(err)
	return
}

func decodeError(src error) (dst error) {
	switch {
	case src == io.EOF:
		dst = src // return as it is
	case status.Code(src) == codes.OK:
		dst = nil
	case status.Code(src) == codes.InvalidArgument:
		dst = fmt.Errorf("%w: %s", ErrInvalid, src)
	default:
		dst = fmt.Errorf("%w: %s", ErrInternal, src)
	}
	return
}
