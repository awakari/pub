package model

import (
	"context"
	"github.com/awakari/pub/api/grpc/ce"
	"io"
)

type MessagesWriter interface {
	io.Closer

	// Write writes the specified messages and returns the accepted count preserving the order.
	// Returns io.EOF if the destination file/connection/whatever is closed.
	Write(ctx context.Context, msgs []*ce.CloudEvent) (ackCount uint32, err error)
}
