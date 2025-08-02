package pub

import (
	"errors"
	"github.com/awakari/pub/api/grpc/publisher"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/service"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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
	counterEvts             *prometheus.CounterVec
	counterAttrs            *prometheus.CounterVec
}

func NewHandler(
	writer publisher.Service,
	writerInternalRateLimit ratelimit.Limiter,
	log *slog.Logger,
	svc service.Service,
	counterEvts *prometheus.CounterVec,
	counterAttrs *prometheus.CounterVec,
) Handler {
	return handler{
		writer:                  writer,
		writerInternalRateLimit: writerInternalRateLimit, // ratelimit.New(writerInternalCfg.RateLimitPerMinute, ratelimit.Per(time.Minute)),
		log:                     log,
		svc:                     svc,
		counterAttrs:            counterAttrs,
		counterEvts:             counterEvts,
	}
}

// Write godoc
// @Summary Publish an event
// @Schemes
// @Description Submit a single CloudEvent to Awakari
// @Tags Events
// @Accept json
// @Param payload body Event true "https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md#32-examples"
// @Param X-Awakari-Group-Id header string true "default"
// @Param X-Awakari-User-Id header string true "foo"
// @Param Authorization	header string true "Bearer XXX..."
// @Success 200 {object} PublishResponse
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "event is blacklisted"
// @Failure 429 {string} string "hourly or daily publishing limit reached"
// @Failure 500 {string} string "internal failure"
// @Failure 503 {string} string "failed to submit, try later"
// @Router /v1 [post]
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

// WriteBatch godoc
// @Summary Publish a batch of events
// @Schemes
// @Description Submit a batch of CloudEvents to Awakari
// @Tags Events
// @Accept json
// @Param payload body EventBatch true "https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md#4-json-batch-format"
// @Param X-Awakari-Group-Id header string true "default"
// @Param X-Awakari-User-Id header string true "foo"
// @Param Authorization	header string true "Bearer XXX..."
// @Success 200 {object} PublishResponse
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "event is blacklisted"
// @Failure 429 {string} string "hourly or daily publishing limit reached"
// @Failure 500 {string} string "internal failure"
// @Failure 503 {string} string "failed to submit, try later"
// @Router /v1/batch [post]
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
		raw, _ := sonic.Marshal(PublishResponse{
			AckCount: resp.AckCount,
		})
		ctx.Data(http.StatusOK, gin.MIMEJSON, raw)
		h.accountConsumedEvents(evts[:resp.AckCount])
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

func (h handler) accountConsumedEvents(evts []*pb.CloudEvent) {
	for _, msg := range evts {
		var lang string
		for k, attr := range msg.Attributes {
			switch attr.Attr.(type) {
			case *pb.CloudEventAttributeValue_CeBoolean:
				h.counterAttrs.WithLabelValues(k, "boolean").Inc()
			case *pb.CloudEventAttributeValue_CeBytes:
				h.counterAttrs.WithLabelValues(k, "bytes").Inc()
			case *pb.CloudEventAttributeValue_CeInteger:
				h.counterAttrs.WithLabelValues(k, "int32").Inc()
			case *pb.CloudEventAttributeValue_CeString:
				h.counterAttrs.WithLabelValues(k, "string").Inc()
				switch k {
				case "language":
					lang = attr.GetCeString()
				}
			case *pb.CloudEventAttributeValue_CeUri:
				h.counterAttrs.WithLabelValues(k, "uri").Inc()
			case *pb.CloudEventAttributeValue_CeUriRef:
				h.counterAttrs.WithLabelValues(k, "uriref").Inc()
			case *pb.CloudEventAttributeValue_CeTimestamp:
				h.counterAttrs.WithLabelValues(k, "timestamp").Inc()
			}
		}
		h.
			counterEvts.
			WithLabelValues(lang, msg.Type).
			Inc()
	}
	return
}
