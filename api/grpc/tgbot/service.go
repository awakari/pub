package tgbot

import (
	"context"
	"github.com/awakari/pub/model"
)

type Service interface {
	ReadChannel(ctx context.Context, link string) (ch *Channel, err error)
	ListChannels(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
}

type service struct {
	client ServiceClient
}

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) ReadChannel(ctx context.Context, link string) (ch *Channel, err error) {
	req := ListChannelsRequest{
		Filter: &Filter{
			Pattern: link,
		},
		Limit: 1000,
	}
	var resp *ListChannelsResponse
	resp, err = svc.client.ListChannels(ctx, &req)
	if err == nil {
		for _, c := range resp.Page {
			if c.Link == link {
				ch = c
			}
		}
	}
	return
}

func (svc service) ListChannels(ctx context.Context, filter *Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	req := ListChannelsRequest{
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
	var resp *ListChannelsResponse
	resp, err = svc.client.ListChannels(ctx, &req)
	if err == nil {
		for _, ch := range resp.Page {
			page = append(page, ch.Link)
		}
	}
	return
}
