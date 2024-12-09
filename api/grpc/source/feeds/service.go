package feeds

import (
	"context"
	"github.com/awakari/pub/model"
)

type Service interface {

	// Create stores a new feed record.
	Create(ctx context.Context, feed *Feed) (msg string, err error)

	Read(ctx context.Context, url string) (feed *Feed, err error)

	Delete(ctx context.Context, url, groupId, userId string) (err error)

	ListUrls(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
}

type service struct {
	client ServiceClient
}

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, feed *Feed) (msg string, err error) {
	var resp *CreateResponse
	resp, err = svc.client.Create(ctx, &CreateRequest{
		Feed: feed,
	})
	if resp != nil {
		msg = resp.Msg
	}
	return
}

func (svc service) Read(ctx context.Context, url string) (feed *Feed, err error) {
	var resp *ReadResponse
	resp, err = svc.client.Read(ctx, &ReadRequest{
		Url: url,
	})
	if resp != nil {
		feed = resp.Feed
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

func (svc service) ListUrls(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
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
