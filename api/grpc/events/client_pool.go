package events

import (
	"context"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
)

type clientPool struct {
	connPool *grpcpool.Pool
}

func NewClientPool(connPool *grpcpool.Pool) ServiceClient {
	return clientPool{
		connPool: connPool,
	}
}

func (cp clientPool) SetStream(ctx context.Context, req *SetStreamRequest, opts ...grpc.CallOption) (resp *SetStreamResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.SetStream(ctx, req, opts...)
	}
	return
}

func (cp clientPool) Publish(ctx context.Context, opts ...grpc.CallOption) (stream Service_PublishClient, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	var c *grpc.ClientConn
	if err == nil {
		c = conn.ClientConn
		conn.Close() // return back to the conn pool immediately
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(c)
		stream, err = client.Publish(ctx, opts...)
	}
	return
}
