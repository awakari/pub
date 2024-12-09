package auth

import (
	"github.com/awakari/pub/api/grpc/source/telegram"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TelegramAuth interface {
	ClientLogin(ginCtx *gin.Context)
}

type tgAuth struct {
	svcTgSrc telegram.Service
}

type LoginFormData struct {
	Code       string `form:"code"`
	ReplicaIdx uint32 `form:"replicaIdx"`
}

func NewTelegramValidator(svcTgSrc telegram.Service) TelegramAuth {
	return tgAuth{
		svcTgSrc: svcTgSrc,
	}
}

func (tgv tgAuth) ClientLogin(ctx *gin.Context) {
	var loginData LoginFormData
	if err := ctx.Bind(&loginData); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		ctx.Abort()
		return
	}
	success, err := tgv.svcTgSrc.Login(ctx, loginData.Code, loginData.ReplicaIdx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		ctx.Abort()
		return
	}
	if !success {
		ctx.Status(http.StatusLocked)
		ctx.Abort()
		return
	}
	ctx.Status(http.StatusAccepted)
	return
}
