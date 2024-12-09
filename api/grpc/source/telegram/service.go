package telegram

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Service interface {
	Create(ctx context.Context, ch *Channel) (err error)
	Read(ctx context.Context, link string) (ch *Channel, err error)
	Delete(ctx context.Context, link string) (err error)
	List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error)

	Login(ctx context.Context, code string, replicaIdx uint32) (success bool, err error)
}

type service struct {
	client        ServiceClient
	fmtUriReplica string
}

func NewService(client ServiceClient, fmtUriReplica string) Service {
	return service{
		client:        client,
		fmtUriReplica: fmtUriReplica,
	}
}

func (svc service) Create(ctx context.Context, ch *Channel) (err error) {
	_, err = svc.client.Create(ctx, &CreateRequest{
		Channel: ch,
	})
	return
}

func (svc service) Read(ctx context.Context, link string) (ch *Channel, err error) {
	var resp *ReadResponse
	resp, err = svc.client.Read(ctx, &ReadRequest{
		Link: link,
	})
	if resp != nil {
		ch = resp.Channel
	}
	return
}

func (svc service) Delete(ctx context.Context, link string) (err error) {
	_, err = svc.client.Delete(ctx, &DeleteRequest{
		Link: link,
	})
	return
}

func (svc service) List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	req := ListRequest{
		Filter: filter,
		Limit:  limit,
		Cursor: cursor,
	}
	switch order {
	case model.OrderDesc:
		req.Order = Order_DESC
	default:
		req.Order = Order_ASC
	}
	var resp *ListResponse
	resp, err = svc.client.List(ctx, &req)
	if err == nil {
		for _, ch := range resp.Page {
			page = append(page, ch.Link)
		}
	}
	return
}

func (svc service) Login(ctx context.Context, code string, replicaIdx uint32) (success bool, err error) {
	replicaUri := fmt.Sprintf(svc.fmtUriReplica, replicaIdx)
	var conn *grpc.ClientConn
	conn, err = grpc.NewClient(replicaUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	var resp *LoginResponse
	if err == nil {
		defer conn.Close()
		client := NewServiceClient(conn)
		resp, err = client.Login(ctx, &LoginRequest{
			Code: code,
		})
	}
	if err == nil {
		success = resp.Success
	}
	return
}
