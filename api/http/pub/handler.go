package pub

import (
	"fmt"
	"github.com/awakari/pub/api/grpc/publisher"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"log/slog"
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
	writerInternalRateLimit ratelimit.Limiter
	blacklist               model.Prefixes[model.BlacklistValue]
	log                     *slog.Logger
}

func NewHandler(
	writer publisher.Service,
	writerInternalCfg config.WriterInternalConfig,
	blacklist model.Prefixes[model.BlacklistValue],
	log *slog.Logger,
) Handler {
	return handler{
		writer:                  writer,
		writerInternalCfg:       writerInternalCfg,
		writerInternalRateLimit: ratelimit.New(writerInternalCfg.RateLimitPerMinute, ratelimit.Per(time.Minute)),
		blacklist:               blacklist,
		log:                     log,
	}
}

func (h handler) Write(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	body, err := io.ReadAll(ctx.Request.Body)
	var evt pb.CloudEvent
	if err == nil {
		err = Unmarshal(body, &evt)

	}
	if err == nil {
		h.write(ctx, []*pb.CloudEvent{&evt}, false)
	}
}

func (h handler) WriteBatch(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	body, err := io.ReadAll(ctx.Request.Body)
	var evts []*pb.CloudEvent
	if err == nil {
		evts, err = UnmarshalBatch(body)
	}
	if err == nil {
		h.write(ctx, evts, false)
	}
}

func (h handler) WriteInternal(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	h.writerInternalRateLimit.Take()
	body, err := io.ReadAll(ctx.Request.Body)
	var evt pb.CloudEvent
	if err == nil {
		err = Unmarshal(body, &evt)
	}
	if err == nil {
		evt.Attributes[h.writerInternalCfg.Name] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: h.writerInternalCfg.Value,
			},
		}
		h.write(ctx, []*pb.CloudEvent{&evt}, true)
	}
}

func (h handler) write(ctx *gin.Context, evts []*pb.CloudEvent, internal bool) {

	if !internal {

		for i, evt := range evts {

			var prefix string
			var attrName string
			var attrValue string
			prefix, _, _ = h.blacklist.FindOnePrefix(ctx, "source:"+evt.Source)
			switch prefix {
			case "":
				prefix, _, _ = h.blacklist.FindOnePrefix(ctx, "type:"+evt.Source)
			default:
				attrName = "source"
				attrValue = evt.Source
			}
			switch prefix {
			case "":
				for k, v := range evt.Attributes {
					switch vt := v.Attr.(type) {
					case *pb.CloudEventAttributeValue_CeString:
						attrValue = vt.CeString
					case *pb.CloudEventAttributeValue_CeUri:
						attrValue = vt.CeUri
					case *pb.CloudEventAttributeValue_CeUriRef:
						attrValue = vt.CeUriRef
					}
					if attrValue != "" {
						prefix, _, _ = h.blacklist.FindOnePrefix(ctx, k+":"+attrValue)
						if prefix != "" {
							attrName = k
							break
						}
					}
				}
			default:
				attrName = "type"
				attrValue = evt.Type
			}

			if prefix != "" {
				switch i {
				case 0:
					h.log.Info(fmt.Sprintf("event was rejected by blacklist prefix: %s, id: %s, attribute: %s=%s\n", prefix, evt.Id, attrName, attrValue))
					ctx.String(http.StatusForbidden, fmt.Sprintf("forbidden by prefix: %s", prefix))
					return
				default:
					evts = evts[:i] // truncate the batch
				}
			}
		}
	}

	grpcCtx, groupId, userId := grpc.AuthRequestContext(ctx)
	for _, evt := range evts {
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
	}
	req := publisher.SubmitMessagesRequest{
		Msgs: evts,
	}
	var resp *publisher.SubmitMessagesResponse
	var err error
	if internal {
		resp, err = h.writer.SubmitInternalEvents(grpcCtx, &req)
	} else {
		resp, err = h.writer.SubmitPermittedEvents(grpcCtx, &req, groupId, userId)
	}

	if err == nil && resp.AckCount == 0 {
		ctx.String(http.StatusServiceUnavailable, "was unable to submit, retry later")
		return
	}

	switch status.Code(err) {
	case codes.OK:
		raw, _ := sonic.Marshal(response{
			AckCount: resp.AckCount,
		})
		ctx.Data(http.StatusOK, gin.MIMEJSON, raw)
	case codes.NotFound:
		ctx.String(http.StatusNotFound, err.Error())
	case codes.AlreadyExists:
		ctx.String(http.StatusConflict, err.Error())
	case codes.Unauthenticated:
		ctx.String(http.StatusUnauthorized, err.Error())
	case codes.DeadlineExceeded:
		ctx.String(http.StatusRequestTimeout, err.Error())
	case codes.InvalidArgument:
		ctx.String(http.StatusBadRequest, err.Error())
	case codes.ResourceExhausted:
		ctx.String(http.StatusTooManyRequests, err.Error())
	case codes.Unavailable:
		ctx.String(http.StatusServiceUnavailable, err.Error())
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}
