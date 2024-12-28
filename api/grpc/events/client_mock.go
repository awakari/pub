package events

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clientMock struct {
}

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (cm clientMock) SetStream(ctx context.Context, req *SetStreamRequest, opts ...grpc.CallOption) (resp *SetStreamResponse, err error) {
	switch req.Topic {
	case "":
		err = status.Error(codes.InvalidArgument, "empty topic")
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	default:
		resp = &SetStreamResponse{}
	}
	return
}

func (cm clientMock) PublishBatch(ctx context.Context, req *PublishRequest, opts ...grpc.CallOption) (resp *PublishResponse, err error) {
	switch req.Topic {
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	default:
		resp = &PublishResponse{
			AckCount: 42,
		}
	}
	return
}
