package auth

import (
	"errors"
	"fmt"
	"github.com/awakari/pub/api/grpc/auth"
	"github.com/awakari/pub/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Handler struct {
	Svc auth.Service
}

func (h Handler) Authorize(ctx *gin.Context) {
	userId := ctx.GetHeader(model.KeyUserId)
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.String(http.StatusUnauthorized, "missing authorization header")
		ctx.Abort()
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		ctx.String(http.StatusUnauthorized, "missing authorization token")
		ctx.Abort()
		return
	}
	err := h.Svc.Authenticate(ctx, userId, token)
	switch {
	case err == nil:
		ctx.Set(model.KeyGroupId, ctx.GetHeader(model.KeyGroupId))
		ctx.Set(model.KeyUserId, userId)
	case errors.Is(err, auth.ErrInvalidToken):
		ctx.String(http.StatusUnauthorized, "invalid token")
		ctx.Abort()
		return
	case errors.Is(err, auth.ErrInvalidUserId):
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid user id: %s", userId))
		ctx.Abort()
		return
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
		ctx.Abort()
		return
	}
	return
}
