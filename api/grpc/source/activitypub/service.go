package activitypub

import (
	"context"
	"github.com/awakari/pub/model"
)

type Service interface {
	Create(ctx context.Context, addr, groupId, userId string) (url string, err error)
	Read(ctx context.Context, url string) (src *Source, err error)
	List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Delete(ctx context.Context, url, groupId, userId string) (err error)
}

type service struct {
	client ServiceClient
}

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, addr, groupId, userId string) (url string, err error) {
	var resp *CreateResponse
	resp, err = svc.client.Create(ctx, &CreateRequest{
		Addr:    addr,
		GroupId: groupId,
		UserId:  userId,
	})
	if resp != nil {
		url = resp.Url
	}
	return
}

func (svc service) Read(ctx context.Context, url string) (src *Source, err error) {
	var resp *ReadResponse
	resp, err = svc.client.Read(ctx, &ReadRequest{
		Url: url,
	})
	if resp != nil {
		src = resp.Src
	}
	return
}

func (svc service) List(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	req := ListUrlsRequest{
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
	var resp *ListUrlsResponse
	resp, err = svc.client.ListUrls(ctx, &req)
	if resp != nil {
		page = resp.Page
	}
	return
}

func (svc service) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	_, err = svc.client.Delete(ctx, &DeleteRequest{
		Url:     url,
		GroupId: groupId,
		UserId:  userId,
	})
	return
}
