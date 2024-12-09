package grpc

import (
	"context"
	grpcAuth "github.com/awakari/pub/api/grpc/auth"
	"github.com/awakari/pub/model"
	"github.com/gin-gonic/gin"
)

func AuthRequestContext(src *gin.Context) (dst context.Context, groupId, userId string) {
	groupId = src.GetString(model.KeyGroupId)
	userId = src.GetString(model.KeyUserId)
	dst = grpcAuth.SetIncomingAuthInfo(src, groupId, userId)
	return
}
