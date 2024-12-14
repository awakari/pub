package pub

import (
	"fmt"
	apiGrpcCe "github.com/awakari/pub/api/grpc/ce"
	"github.com/awakari/pub/api/grpc/events"
	"github.com/awakari/pub/api/grpc/publisher"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/gin-gonic/gin"
	grpcpool "github.com/processout/grpc-go-pool"
	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"time"
)

type Handler interface {
	Write(ctx *gin.Context)
	WriteBatch(ctx *gin.Context)
	WriteInternal(ctx *gin.Context)
}

type handler struct {
	writer                  publisher.Service
	writerInternalCfg       config.WriterInternalConfig
	connPoolEvts            *grpcpool.Pool
	writerInternalRateLimit ratelimit.Limiter
	blacklist               model.Prefixes[model.BlacklistValue]
}

func NewHandler(
	writer publisher.Service,
	writerInternalCfg config.WriterInternalConfig,
	connPoolEvts *grpcpool.Pool,
	blacklist model.Prefixes[model.BlacklistValue],
) Handler {
	return handler{
		writer:                  writer,
		writerInternalCfg:       writerInternalCfg,
		connPoolEvts:            connPoolEvts,
		writerInternalRateLimit: ratelimit.New(writerInternalCfg.RateLimitPerMinute, ratelimit.Per(time.Minute)),
		blacklist:               blacklist,
	}
}

func (h handler) Write(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	var evt apiGrpcCe.CloudEvent
	if err == nil {
		err = protojson.Unmarshal(body, &evt)
	}
	if err == nil {
		h.write(ctx, []*apiGrpcCe.CloudEvent{&evt}, false)
	}
}

func (h handler) WriteBatch(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	var evts apiGrpcCe.CloudEventBatch
	if err == nil {
		err = protojson.Unmarshal(body, &evts)
	}
	if err == nil {
		h.write(ctx, evts.Events, false)
	}
}

func (h handler) WriteInternal(ctx *gin.Context) {
	h.writerInternalRateLimit.Take()
	body, err := io.ReadAll(ctx.Request.Body)
	var evt apiGrpcCe.CloudEvent
	if err == nil {
		err = protojson.Unmarshal(body, &evt)
	}
	if err == nil {
		evt.Attributes[h.writerInternalCfg.Name] = &apiGrpcCe.CloudEventAttributeValue{
			Attr: &apiGrpcCe.CloudEventAttributeValue_CeInteger{
				CeInteger: h.writerInternalCfg.Value,
			},
		}
		h.write(ctx, []*apiGrpcCe.CloudEvent{&evt}, true)
	}
}

func (h handler) write(ctx *gin.Context, evts []*apiGrpcCe.CloudEvent, internal bool) {

	if !internal {

		for i, evt := range evts {

			var prefix string
			prefix, _, _ = h.blacklist.FindOnePrefix(ctx, "source:"+evt.Source)
			if prefix == "" {
				prefix, _, _ = h.blacklist.FindOnePrefix(ctx, "type:"+evt.Source)
			}
			if prefix == "" {
				for k, v := range evt.Attributes {
					var vs string
					switch vt := v.Attr.(type) {
					case *apiGrpcCe.CloudEventAttributeValue_CeString:
						vs = vt.CeString
					case *apiGrpcCe.CloudEventAttributeValue_CeUri:
						vs = vt.CeUri
					case *apiGrpcCe.CloudEventAttributeValue_CeUriRef:
						vs = vt.CeUriRef
					}
					if vs != "" {
						prefix, _, _ = h.blacklist.FindOnePrefix(ctx, k+":"+vs)
						if prefix != "" {
							break
						}
					}
				}
			}

			if prefix != "" {
				switch i {
				case 0:
					fmt.Printf("event %s was rejected by blacklist prefix: %s\n", evt.Id, prefix)
					ctx.String(http.StatusForbidden, fmt.Sprintf("forbidden by prefix: %s", prefix))
					return
				default:
					evts = evts[:i] // truncate the batch
				}
			}
		}
	}

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
		grpcCtx, groupId, userId := grpc.AuthRequestContext(ctx)
		for _, evt := range evts {
			if evt.Attributes == nil {
				evt.Attributes = make(map[string]*apiGrpcCe.CloudEventAttributeValue)
			}
			evt.Attributes[model.KeyCeGroupId] = &apiGrpcCe.CloudEventAttributeValue{
				Attr: &apiGrpcCe.CloudEventAttributeValue_CeString{
					CeString: groupId,
				},
			}
			evt.Attributes[model.KeyCeUserId] = &apiGrpcCe.CloudEventAttributeValue{
				Attr: &apiGrpcCe.CloudEventAttributeValue_CeString{
					CeString: userId,
				},
			}
			evt.Attributes[model.KeyCePubTime] = &apiGrpcCe.CloudEventAttributeValue{
				Attr: &apiGrpcCe.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(time.Now().UTC()),
				},
			}
		}
		req := publisher.SubmitMessagesRequest{
			Msgs: evts,
		}
		if internal {
			resp, err = h.writer.SubmitInternalEvents(grpcCtx, streamClient, &req)
		} else {
			resp, err = h.writer.SubmitPermittedEvents(grpcCtx, streamClient, &req, groupId, userId)
		}
	}

	if err == nil && resp.AckCount == 0 {
		ctx.String(http.StatusServiceUnavailable, "was unable to submit, retry later")
		return
	}

	grpc.RespondJson(ctx, resp, err)
}
