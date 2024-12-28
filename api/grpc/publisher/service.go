package publisher

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/pub/api/grpc/events"
	"github.com/awakari/pub/api/grpc/limits"
	"github.com/awakari/pub/api/grpc/permits"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	SubmitPermittedEvents(ctx context.Context, req *SubmitMessagesRequest, groupId, userId string) (resp *SubmitMessagesResponse, err error)
	SubmitInternalEvents(ctx context.Context, req *SubmitMessagesRequest) (resp *SubmitMessagesResponse, err error)
}

type svc struct {
	client     events.ServiceClient
	svcPermits permits.Service
	cfgEvts    config.EventsConfig
}

const subjPubMsgs = model.SubjectPublishEvents
const txtLimitReached = `⚠ Publishing limit reached.

Increase your publishing limit or nominate own sources for the dedicated limit.

If you did not publish messages, <a href="https://awakari.com/pub.html?own=true">check own publication sources</a> you added.`

func NewService(client events.ServiceClient, svcPermits permits.Service, cfgEvts config.EventsConfig) Service {
	return svc{
		client:     client,
		svcPermits: svcPermits,
		cfgEvts:    cfgEvts,
	}
}

func (s svc) SubmitPermittedEvents(ctx context.Context, req *SubmitMessagesRequest, groupId, userId string) (resp *SubmitMessagesResponse, err error) {
	// allocate permit
	var permit model.Permit
	permit, err = s.svcPermits.Request(ctx, groupId, userId, subjPubMsgs, uint32(len(req.Msgs)))
	err = encodeError(err)
	// utilize permit
	if err == nil {
		resp, err = s.utilizePermit(ctx, req, permit, groupId)
	}
	var usedCount uint32
	if err == nil {
		usedCount = resp.AckCount
	}
	// release the unused permit count
	unusedCount := permit.Count - usedCount
	if unusedCount > 0 {
		_ = s.svcPermits.Release(ctx, groupId, permit.UserId, subjPubMsgs, unusedCount)
	}
	return
}

func (s svc) SubmitInternalEvents(ctx context.Context, req *SubmitMessagesRequest) (resp *SubmitMessagesResponse, err error) {
	// proxy a request
	var dstResp *events.PublishResponse
	dstResp, err = s.client.PublishBatch(ctx, &events.PublishRequest{
		Topic: s.cfgEvts.Topic,
		Evts:  req.Msgs,
	})
	if dstResp != nil {
		resp = &SubmitMessagesResponse{
			AckCount: dstResp.AckCount,
		}
	}
	return
}

func (s svc) utilizePermit(ctx context.Context, srcReq *SubmitMessagesRequest, permit model.Permit, groupId string) (resp *SubmitMessagesResponse, err error) {
	// send the message if permit is just exhausted for the 1st time since last reset
	if permit.JustExhausted {
		resp, err = s.notifyLimitReached(ctx, srcReq, groupId, permit.UserId)
	}
	//
	var dstReq *events.PublishRequest
	if err == nil {
		dstReq, err = s.applyPermit(srcReq, permit)
	}
	// proxy a request
	var dstResp *events.PublishResponse
	if err == nil {
		dstResp, err = s.client.PublishBatch(ctx, dstReq)
	}
	if dstResp != nil {
		resp = &SubmitMessagesResponse{
			AckCount: dstResp.AckCount,
		}
	}
	return
}

func (s svc) notifyLimitReached(
	ctx context.Context,
	srcReq *SubmitMessagesRequest,
	groupId, userId string,
) (
	resp *SubmitMessagesResponse,
	err error,
) {
	var src string
	if srcReq != nil && len(srcReq.Msgs) > 0 {
		src = srcReq.Msgs[0].Source
	}
	if userId == "" || userId == src {
		return // no limit owner, don't send anything
	}
	evt := pb.CloudEvent{
		Attributes: map[string]*pb.CloudEventAttributeValue{
			model.KeyToGroupId: {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: groupId,
				},
			},
			model.KeyToUserId: {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: userId,
				},
			},
		},
		Data: &pb.CloudEvent_TextData{
			TextData: txtLimitReached,
		},
		Id:          ksuid.New().String(),
		Source:      src,
		SpecVersion: "1.0",
		Type:        model.ValTypeLimitReached,
	}
	req := SubmitMessagesRequest{
		Msgs: []*pb.CloudEvent{
			&evt,
		},
	}
	resp, err = s.SubmitInternalEvents(ctx, &req)
	fmt.Printf("user %s daily publishing limit reached notification %s status: %+v, %s\n", userId, evt.Id, resp, err)
	return
}

func (s svc) applyPermit(srcReq *SubmitMessagesRequest, permit model.Permit) (dstReq *events.PublishRequest, err error) {
	switch permit.Count {
	case 0:
		err = status.Error(codes.ResourceExhausted, fmt.Sprintf("user id %s: usage limit reached/not set", permit.UserId))
	case uint32(len(srcReq.Msgs)):
		dstReq = &events.PublishRequest{
			Topic: s.cfgEvts.Topic,
			Evts:  srcReq.Msgs,
		}
	default:
		dstReq = &events.PublishRequest{
			Topic: s.cfgEvts.Topic,
			Evts:  srcReq.Msgs[:permit.Count],
		}
	}
	return
}

func encodeError(src error) (dst error) {
	switch {
	case src == nil:
	case errors.Is(src, limits.ErrInternal):
		dst = status.Error(codes.Internal, fmt.Sprintf("limits %s", src))
	case errors.Is(src, permits.ErrInternal):
		dst = status.Error(codes.Internal, fmt.Sprintf("permits %s", src))
	case status.Code(src) != codes.Unknown:
		dst = src // proxy when src error is grpc status
	default:
		dst = status.Error(codes.Unknown, src.Error())
	}
	return
}
