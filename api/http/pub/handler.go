package pub

import (
	"errors"
	"github.com/awakari/pub/api/grpc/publisher"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/service"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
)

type Handler interface {
	Write(ctx *gin.Context)
	WriteBatch(ctx *gin.Context)
	WriteInternal(ctx *gin.Context)
}

type handler struct {
	writer                  publisher.Service
	writerInternalRateLimit ratelimit.Limiter
	log                     *slog.Logger
	svc                     service.Service
}

func NewHandler(
	writer publisher.Service,
	writerInternalRateLimit ratelimit.Limiter,
	log *slog.Logger,
	svc service.Service,
) Handler {
	return handler{
		writer:                  writer,
		writerInternalRateLimit: writerInternalRateLimit, // ratelimit.New(writerInternalCfg.RateLimitPerMinute, ratelimit.Per(time.Minute)),
		log:                     log,
		svc:                     svc,
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
		h.write(ctx, []*pb.CloudEvent{&evt}, true)
	}
}

func (h handler) write(ctx *gin.Context, evts []*pb.CloudEvent, internal bool) {

	grpcCtx, groupId, userId := grpc.AuthRequestContext(ctx)
	for i, evt := range evts {
		err := h.svc.Preprocess(ctx, evt, groupId, userId, internal)
		switch {
		case errors.Is(err, service.ErrRejected):
			switch i {
			case 0:
				h.log.Info(err.Error())
				ctx.String(http.StatusForbidden, err.Error())
				return
			default:
				evts = evts[:i] // truncate the batch
			}
		}
		if err != nil {
			break
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
