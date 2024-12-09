package model

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"io"
)

type MessagesWriter interface {
	io.Closer

	// Write writes the specified messages and returns the accepted count preserving the order.
	// Returns io.EOF if the destination file/connection/whatever is closed.
	Write(ctx context.Context, msgs []*pb.CloudEvent) (ackCount uint32, err error)
}
