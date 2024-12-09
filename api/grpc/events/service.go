package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/pub/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type Service interface {
	SetStream(ctx context.Context, topic string, limit uint32) (err error)
	NewPublisher(ctx context.Context, topic string) (p model.MessagesWriter, err error)
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

func (svc service) NewPublisher(ctx context.Context, topic string) (p model.MessagesWriter, err error) {
	var stream Service_PublishClient
	stream, err = svc.client.Publish(ctx)
	if err == nil {
		p = newPublisher(stream, topic)
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
