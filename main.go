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
	v2 "github.com/awakari/pub/api/http/pub"
	httpSrc "github.com/awakari/pub/api/http/pub/src"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/awakari/pub/service"
	"github.com/awakari/pub/storage"
	"github.com/drankou/go-vader/vader"
	"github.com/gin-gonic/gin"
	"github.com/pebbe/textcat"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/awakari/pub/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	//_ "net/http/pprof"
	"os"
)

// @title           Awakari Publishing API
// @version         1.0
// @description     Publish API service is responsible for publishing messages and managing sources.

// @contact.name   Awakari Support
// @contact.email  awakari@awakari.com

// @BasePath  /

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
	for i := 0; i < cfg.Api.Events.Topics.Count; i++ {
		err = svcEvts.SetStream(context.TODO(), fmt.Sprintf(cfg.Api.Events.Topics.Fmt, i), cfg.Api.Events.Limit)
		if err != nil {
			panic(err)
		}
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
	blacklist = model.NewPrefixesLogging(blacklist, log)
	var blacklistSize uint32
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
			blacklistSize++
		}
	}
	stor.Close()
	log.Info(fmt.Sprintf("loaded the blacklist, size: %d", blacklistSize))

	var sia *vader.SentimentIntensityAnalyzer
	var txtCat *textcat.TextCat
	if cfg.Preprocess.Snippets.Enabled {
		if cfg.Preprocess.Sentiments.Enabled {
			sia = &vader.SentimentIntensityAnalyzer{}
			err = sia.Init()
			if err != nil {
				panic(err)
			}
		}
		txtCat = textcat.NewTextCat()
		txtCat.EnableAllRawLanguages()
		txtCat.EnableAllUtf8Languages()
	}
	svc := service.New(blacklist, cfg.Preprocess, cfg.Api.Writer.Internal, sia, txtCat)

	counterEvts := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "awk_published_events_count",
			Help: "The total number of events submitted to Awakari",
		},
		[]string{
			"language",
			"type",
		},
	)
	counterAttrs := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "awk_published_attrs_observed_count",
			Help: "The total number of event attributes observed Awakari resolver",
		},
		[]string{
			"key",
			"type",
		},
	)

	handlerPub := v2.NewHandler(
		publisher.NewService(clientEvts, svcPermits, cfg.Api.Events),
		ratelimit.New(cfg.Api.Writer.Internal.RateLimitPerMinute, ratelimit.Per(time.Minute)),
		log,
		svc,
		counterEvts,
		counterAttrs,
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

	// expose the profiling
	//go func() {
	//    _ = http.ListenAndServe("localhost:6060", nil)
	//}()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", cfg.Api.Metrics.Port), nil)

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
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
