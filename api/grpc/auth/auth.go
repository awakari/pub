package auth

import (
	"context"
	"fmt"
	"github.com/awakari/pub/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func SetOutgoingAuthInfo(src context.Context, groupId, userId string) (dst context.Context) {
	dst = metadata.AppendToOutgoingContext(src, model.KeyGroupId, groupId, model.KeyUserId, userId)
	return
}

func SetIncomingAuthInfo(src context.Context, groupId, userId string) (dst context.Context) {
	md := metadata.MD{
		model.KeyGroupId: []string{
			groupId,
		},
		model.KeyUserId: []string{
			userId,
		},
	}
	dst = metadata.NewIncomingContext(src, md)
	return
}

func GetIncomingAuthInfo(ctx context.Context) (groupId, userId string, err error) {
	groupId, userId, err = getIncomingAuthInfo(ctx, true)
	return
}

func getIncomingAuthInfo(ctx context.Context, userIdRequired bool) (groupId, userId string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Error(codes.Unauthenticated, "missing request metadata")
	}
	if err == nil {
		groupId = getMetadataValue(md, model.KeyGroupId)
		if groupId == "" {
			err = status.Error(codes.Unauthenticated, fmt.Sprintf("missing value for %s in request metadata", model.KeyGroupId))
		}
	}
	if err == nil {
		userId = getMetadataValue(md, model.KeyUserId)
		if userIdRequired && userId == "" {
			err = status.Error(codes.Unauthenticated, fmt.Sprintf("missing value for %s in request metadata", model.KeyUserId))
		}
	}
	return
}

func getMetadataValue(md metadata.MD, k string) (v string) {
	var vals []string
	if vals = md.Get(k); len(vals) > 0 {
		v = vals[0]
	}
	return
}
