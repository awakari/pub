package permits

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

func (cp clientPool) Get(ctx context.Context, req *GetRequest, opts ...grpc.CallOption) (resp *GetResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.Get(ctx, req, opts...)
	}
	return
}

func (cp clientPool) Allocate(ctx context.Context, req *AllocateRequest, opts ...grpc.CallOption) (resp *AllocateResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.Allocate(ctx, req, opts...)
	}
	return
}

func (cp clientPool) Release(ctx context.Context, req *ReleaseRequest, opts ...grpc.CallOption) (resp *ReleaseResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.Release(ctx, req, opts...)
	}
	return
}
