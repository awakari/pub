package pub

import (
	"github.com/awakari/pub/api/grpc/events"
	"github.com/awakari/pub/api/grpc/publisher"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/gin-gonic/gin"
	grpcpool "github.com/processout/grpc-go-pool"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

type Handler interface {
	Write(ctx *gin.Context)
	WriteInternal(ctx *gin.Context)
}

type handler struct {
	writer                  publisher.Service
	writerInternalCfg       config.WriterInternalConfig
	connPoolEvts            *grpcpool.Pool
	writerInternalRateLimit ratelimit.Limiter
}

func NewHandler(writer publisher.Service, writerInternalCfg config.WriterInternalConfig, connPoolEvts *grpcpool.Pool) Handler {
	return handler{
		writer:                  writer,
		writerInternalCfg:       writerInternalCfg,
		connPoolEvts:            connPoolEvts,
		writerInternalRateLimit: ratelimit.New(writerInternalCfg.RateLimitPerMinute, ratelimit.Per(time.Minute)),
	}
}

func (h handler) Write(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	var evt pb.CloudEvent
	if err == nil {
		err = protojson.Unmarshal(body, &evt)
	}
	if err == nil {
		h.write(ctx, &evt, false)
	}
}

func (h handler) WriteInternal(ctx *gin.Context) {
	h.writerInternalRateLimit.Take()
	body, err := io.ReadAll(ctx.Request.Body)
	var evt pb.CloudEvent
	if err == nil {
		err = protojson.Unmarshal(body, &evt)
	}
	if err == nil {
		evt.Attributes[h.writerInternalCfg.Name] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: h.writerInternalCfg.Value,
			},
		}
		h.write(ctx, &evt, true)
	}
}

func (h handler) write(ctx *gin.Context, evt *pb.CloudEvent, internal bool) {
	grpcCtx, groupId, userId := grpc.AuthRequestContext(ctx)
	conn, err := h.connPoolEvts.Get(ctx)
	var streamClient events.Service_PublishClient
	if err == nil {
		c := conn.ClientConn
		conn.Close() // return back to the conn pool immediately
		client := events.NewServiceClient(c)
		streamClient, err = client.Publish(ctx)
	}
	var resp *publisher.SubmitMessagesResponse
	if err == nil {
		defer streamClient.CloseSend()
		if evt.Attributes == nil {
			evt.Attributes = make(map[string]*pb.CloudEventAttributeValue)
		}
		evt.Attributes[model.KeyCeGroupId] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: groupId,
			},
		}
		evt.Attributes[model.KeyCeUserId] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: userId,
			},
		}
		evt.Attributes[model.KeyCePubTime] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(time.Now().UTC()),
			},
		}
		req := publisher.SubmitMessagesRequest{
			Msgs: []*pb.CloudEvent{
				evt,
			},
		}
		if internal {
			resp, err = h.writer.SubmitInternalEvents(grpcCtx, streamClient, &req)
		} else {
			resp, err = h.writer.SubmitPermittedEvents(grpcCtx, streamClient, &req, groupId, userId)
		}
	}
	if err == nil && resp.AckCount == 0 {
		err = status.Error(codes.Unavailable, "was unable to submit, retry later")
	}
	grpc.RespondJson(ctx, resp, err)
}
