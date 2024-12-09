package grpc

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

var protoJsonMarshalOpts = protojson.MarshalOptions{
	Multiline:      true,
	Indent:         "  ",
	AllowPartial:   true,
	UseProtoNames:  true,
	UseEnumNumbers: true,
}

func RespondJson(ctx *gin.Context, resp proto.Message, err error) {
	switch status.Code(err) {
	case codes.OK:
		raw, _ := protoJsonMarshalOpts.Marshal(resp)
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
