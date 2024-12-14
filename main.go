package main

import (
	"context"
	"fmt"
	grpcAuth "github.com/awakari/pub/api/grpc/auth"
	"github.com/awakari/pub/api/grpc/events"
	grpcLimits "github.com/awakari/pub/api/grpc/limits"
	grpcPermits "github.com/awakari/pub/api/grpc/permits"
	"github.com/awakari/pub/api/grpc/publisher"
	grpcSrcAp "github.com/awakari/pub/api/grpc/source/activitypub"
	grpcSrcFeeds "github.com/awakari/pub/api/grpc/source/feeds"
	grpcSrcSites "github.com/awakari/pub/api/grpc/source/sites"
	grpcSrcTg "github.com/awakari/pub/api/grpc/source/telegram"
	"github.com/awakari/pub/api/grpc/tgbot"
	auth2 "github.com/awakari/pub/api/http/auth"
	httpPub "github.com/awakari/pub/api/http/pub"
	httpSrc "github.com/awakari/pub/api/http/pub/src"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/awakari/pub/storage"
	"github.com/gin-gonic/gin"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
)

func main() {
	//
	slog.Info("starting...")
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load the config: %s", err.Error()))
	}
	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))

	connPoolEvts, err := grpcpool.New(
		func() (*grpc.ClientConn, error) {
			return grpc.NewClient(cfg.Api.Events.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
		},
		int(cfg.Api.Events.Connection.Count.Init),
		int(cfg.Api.Events.Connection.Count.Max),
		cfg.Api.Events.Connection.IdleTimeout,
	)
	if err != nil {
		panic(err)
	}
	defer connPoolEvts.Close()
	clientEvts := events.NewClientPool(connPoolEvts)
	svcEvts := events.NewService(clientEvts)
	svcEvts = events.NewLoggingMiddleware(svcEvts, log)
	err = svcEvts.SetStream(context.TODO(), cfg.Api.Events.Topic, cfg.Api.Events.Limit)
	if err != nil {
		panic(err)
	}

	// init the source-feeds client
	connSrcFeeds, err := grpc.NewClient(cfg.Api.Source.Feeds.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		log.Info("connected the source-feeds service")
		defer connSrcFeeds.Close()
	} else {
		log.Error(fmt.Sprintf("failed to connect the source-feeds service: %s", err))
	}
	clientSrcFeeds := grpcSrcFeeds.NewServiceClient(connSrcFeeds)
	svcSrcFeeds := grpcSrcFeeds.NewService(clientSrcFeeds)
	svcSrcFeeds = grpcSrcFeeds.NewServiceLogging(svcSrcFeeds, log)

	// init the source-telegram client
	connSrcTg, err := grpc.NewClient(cfg.Api.Source.Telegram.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		log.Info("connected the source-telegram service")
		defer connSrcTg.Close()
	} else {
		log.Error(fmt.Sprintf("failed to connect the source-telegram service: %s", err))
	}
	clientSrcTg := grpcSrcTg.NewServiceClient(connSrcTg)
	svcSrcTg := grpcSrcTg.NewService(clientSrcTg, cfg.Api.Source.Telegram.FmtUriReplica)
	svcSrcTg = grpcSrcTg.NewServiceLogging(svcSrcTg, log)

	// init the source-sites client
	connSrcSites, err := grpc.NewClient(cfg.Api.Source.Sites.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		log.Info("connected the source-sites service")
		defer connSrcSites.Close()
	} else {
		log.Error(fmt.Sprintf("failed to connect the source-sites service: %s", err))
	}
	clientSrcSites := grpcSrcSites.NewServiceClient(connSrcSites)
	svcSrcSites := grpcSrcSites.NewService(clientSrcSites)
	svcSrcSites = grpcSrcSites.NewServiceLogging(svcSrcSites, log)

	// init the int-activitypub client
	connSrcAp, err := grpc.NewClient(cfg.Api.Source.ActivityPub.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		log.Info("connected the int-activitypub service")
		defer connSrcAp.Close()
	} else {
		log.Error(fmt.Sprintf("failed to connect the int-activitypub service: %s", err))
	}
	clientSrcAp := grpcSrcAp.NewServiceClient(connSrcAp)
	svcSrcAp := grpcSrcAp.NewService(clientSrcAp)
	svcSrcAp = grpcSrcAp.NewLogging(svcSrcAp, log)

	connTgBot, err := grpc.NewClient(cfg.Api.TgBot.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		log.Info("connected the the bot-telegram service")
	} else {
		log.Warn(fmt.Sprintf("failed to connect the bot-telegram: %s", err))
	}
	var clientTgBot tgbot.ServiceClient
	if connTgBot != nil {
		clientTgBot = tgbot.NewServiceClient(connTgBot)
	}
	var svcTgBot tgbot.Service
	if clientTgBot != nil {
		svcTgBot = tgbot.NewService(clientTgBot)
		svcTgBot = tgbot.NewServiceLogging(svcTgBot, log)
	}

	connPoolLimits, err := grpcpool.New(
		func() (*grpc.ClientConn, error) {
			return grpc.NewClient(cfg.Api.Usage.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
		},
		int(cfg.Api.Usage.Connection.Count.Init),
		int(cfg.Api.Usage.Connection.Count.Max),
		cfg.Api.Usage.Connection.IdleTimeout,
	)
	if err != nil {
		panic(err)
	}
	defer connPoolLimits.Close()
	clientLimits := grpcLimits.NewClientPool(connPoolLimits)
	svcLimits := grpcLimits.NewService(clientLimits)
	svcLimits = grpcLimits.NewServiceLogging(svcLimits, log)

	connPoolPermits, err := grpcpool.New(
		func() (*grpc.ClientConn, error) {
			return grpc.NewClient(cfg.Api.Usage.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
		},
		int(cfg.Api.Usage.Connection.Count.Init),
		int(cfg.Api.Usage.Connection.Count.Max),
		cfg.Api.Usage.Connection.IdleTimeout,
	)
	if err != nil {
		panic(err)
	}
	defer connPoolLimits.Close()
	clientPermits := grpcPermits.NewClientPool(connPoolPermits)
	svcPermits := grpcPermits.NewService(clientPermits)
	svcPermits = grpcPermits.NewServiceLogging(svcPermits, log)

	// init blacklist
	var stor storage.Blacklist
	stor, err = storage.NewBlacklist(context.TODO(), cfg.Db)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the blacklist storage: %s", err))
	}
	var cursor string
	var page []model.BlacklistEntry
	blacklist := model.NewPrefixes[model.BlacklistValue]()
	for {
		page, err = stor.GetPage(context.TODO(), 100, cursor)
		if err != nil {
			panic(err)
		}
		if len(page) == 0 {
			break
		}
		cursor = page[len(page)-1].Prefix
		for _, e := range page {
			_ = blacklist.Put(context.TODO(), e.Prefix, e.Value)
		}
	}
	stor.Close()
	log.Info("loaded the blacklist")

	handlerPub := httpPub.NewHandler(
		publisher.NewService(clientEvts, svcPermits, cfg.Api.Events),
		cfg.Api.Writer.Internal,
		connPoolEvts,
		blacklist,
	)
	handlerSrc := httpSrc.NewHandler(svcSrcFeeds, svcSrcSites, svcSrcTg, svcSrcAp, svcTgBot, svcLimits, svcPermits)

	connAuth, err := grpc.NewClient(cfg.Api.Auth.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	clientAuth := grpcAuth.NewServiceClient(connAuth)
	svcAuth := grpcAuth.NewService(clientAuth)
	svcAuth = grpcAuth.NewLogging(svcAuth, log)
	handlerAuth := auth2.Handler{
		Svc: svcAuth,
	}

	authSrcTg := auth2.NewTelegramValidator(svcSrcTg)

	r := gin.Default()
	r.
		Group("/v1/src/:type").
		POST("", handlerAuth.Authorize, handlerSrc.Create).
		GET("", handlerAuth.Authorize, handlerSrc.Read).
		DELETE("", handlerAuth.Authorize, handlerSrc.Delete).
		GET("/list", handlerAuth.Authorize, handlerSrc.List)
	r.
		Group("/v1/tg", handlerAuth.Authorize).
		POST("", authSrcTg.ClientLogin)
	r.
		Group("/v1", handlerAuth.Authorize).
		POST("", handlerPub.Write).
		POST("/batch", handlerPub.WriteBatch).
		POST("/internal", handlerPub.WriteInternal)
	err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
	if err != nil {
		panic(err)
	}
}
