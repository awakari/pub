package limits

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

func (cp clientPool) Set(ctx context.Context, req *SetRequest, opts ...grpc.CallOption) (resp *SetResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.Set(ctx, req, opts...)
	}
	return
}

func (cp clientPool) Delete(ctx context.Context, req *DeleteRequest, opts ...grpc.CallOption) (resp *DeleteResponse, err error) {
	var conn *grpcpool.ClientConn
	conn, err = cp.connPool.Get(ctx)
	if err == nil {
		defer conn.Close()
	}
	var client ServiceClient
	if err == nil {
		client = NewServiceClient(conn)
		resp, err = client.Delete(ctx, req, opts...)
	}
	return
}
