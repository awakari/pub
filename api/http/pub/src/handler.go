package src

import (
	"fmt"
	"github.com/awakari/pub/api/grpc/limits"
	"github.com/awakari/pub/api/grpc/permits"
	"github.com/awakari/pub/api/grpc/source/activitypub"
	"github.com/awakari/pub/api/grpc/source/feeds"
	"github.com/awakari/pub/api/grpc/source/sites"
	"github.com/awakari/pub/api/grpc/source/telegram"
	"github.com/awakari/pub/api/grpc/tgbot"
	"github.com/awakari/pub/api/http/grpc"
	"github.com/awakari/pub/model"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Handler interface {
	Create(ctx *gin.Context)
	Read(ctx *gin.Context)
	Delete(ctx *gin.Context)
	List(ctx *gin.Context)
}

type handler struct {
	svcFeeds   feeds.Service
	svcSites   sites.Service
	svcTg      telegram.Service
	svcAp      activitypub.Service
	svcTgBot   tgbot.Service
	svcLimits  limits.Service
	svcPermits permits.Service
}

const day = 24 * time.Hour
const pageLimitDefault = 10
const keySrcAddr = "X-Awakari-Src-Addr"

func NewHandler(
	svcFeeds feeds.Service,
	svcSites sites.Service,
	svcTg telegram.Service,
	svcAp activitypub.Service,
	svcTgBot tgbot.Service,
	svcLimits limits.Service,
	svcPermits permits.Service,
) Handler {
	return handler{
		svcFeeds:   svcFeeds,
		svcSites:   svcSites,
		svcTg:      svcTg,
		svcAp:      svcAp,
		svcTgBot:   svcTgBot,
		svcLimits:  svcLimits,
		svcPermits: svcPermits,
	}
}

func (h handler) Create(ctx *gin.Context) {
	_, groupId, userId := grpc.AuthRequestContext(ctx)
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	_ = ctx.Request.Body.Close()
	var payload CreatePayload
	err = sonic.Unmarshal(body, &payload)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	payload.Src.Type = ctx.Param("type")
	err = payload.validate()
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	var msg string
	switch payload.Src.Type {
	case TypeApub:
		msg, err = h.svcAp.Create(ctx, payload.Src.Addr, groupId, userId)
	case TypeFeed:
		msg, err = h.svcFeeds.Create(ctx, &feeds.Feed{
			Url:          payload.Src.Addr,
			GroupId:      groupId,
			UserId:       userId,
			UpdatePeriod: durationpb.New(day / time.Duration(payload.Limit.Freq)),
			NextUpdate:   timestamppb.New(time.Now().UTC()),
		})
	case TypeTgCh:
		err = h.svcTg.Create(ctx, &telegram.Channel{
			GroupId: groupId,
			UserId:  userId,
			Link:    payload.Src.Addr,
		})
	default:
		ctx.String(http.StatusBadRequest, fmt.Sprintf("unsupported source type: %s", payload.Src.Type))
	}
	switch {
	case err == nil:
		ctx.String(http.StatusCreated, msg)
	case status.Code(err) == codes.InvalidArgument:
		ctx.String(http.StatusBadRequest, err.Error())
	case status.Code(err) == codes.AlreadyExists:
		ctx.String(http.StatusConflict, err.Error())
	case status.Code(err) == codes.PermissionDenied:
		ctx.String(http.StatusForbidden, err.Error())
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	return
}

func (h handler) Read(ctx *gin.Context) {
	addrEnc := ctx.GetHeader(keySrcAddr)
	if addrEnc == "" {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("missing header: %s", keySrcAddr))
		return
	}
	addr, err := url.QueryUnescape(addrEnc)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid url encoded source address: %s", addrEnc))
		return
	}
	addr = escapeNonAsciiChars(addr)
	fmt.Printf("read the source by address: %s => %s\n", addrEnc, addr)
	typ := ctx.Param("type")
	var result ReadPayload
	result.Counts = make(map[uint32]int64)
	switch typ {
	case TypeApub:
		var src *activitypub.Source
		src, err = h.svcAp.Read(ctx, addr)
		if err == nil {
			result.Addr = addr
			result.GroupId = src.GroupId
			result.UserId = src.UserId
			result.Push = true
			result.Name = src.Name
			result.Accepted = src.Accepted
			if src.Last != nil {
				result.LastUpdate = src.Last.AsTime()
			}
			if src.Created != nil {
				result.Created = src.Created.AsTime()
			}
			result.Query = src.Term
		}
	case TypeFeed:
		var feed *feeds.Feed
		feed, err = h.svcFeeds.Read(ctx, addr)
		if err == nil {
			result.Addr = addr
			result.GroupId = feed.GroupId
			result.UserId = feed.UserId
			if feed.ItemLast != nil {
				result.LastUpdate = feed.ItemLast.AsTime()
			}
			if feed.UpdatePeriod != nil {
				result.UpdatePeriod = feed.UpdatePeriod.AsDuration()
			}
			if feed.NextUpdate != nil {
				result.NextUpdate = feed.NextUpdate.AsTime()
			}
			result.Push = feed.Push
			for o, c := range feed.Counts {
				result.Counts[o] = c
			}
			result.Accepted = true
			if feed.Created != nil {
				result.Created = feed.Created.AsTime()
			}
			result.Query = feed.Terms
			result.Name = feed.Title
		}
	case TypeSite:
		var site *sites.Site
		site, err = h.svcSites.Read(ctx, addr)
		if err == nil {
			result.Addr = addr
			result.GroupId = site.GroupId
			result.UserId = site.UserId
			if site.LastUpdate != nil {
				result.LastUpdate = site.LastUpdate.AsTime()
			}
			result.Accepted = true
		}
	case TypeTgCh:
		var ch *telegram.Channel
		ch, err = h.svcTg.Read(ctx, addr)
		if err == nil {
			result.Addr = addr
			result.GroupId = ch.GroupId
			result.UserId = ch.UserId
			result.Push = true
			result.Name = ch.Name
			result.Accepted = true
			if ch.Created != nil {
				result.Created = ch.Created.AsTime()
			}
			if ch.Last != nil {
				result.LastUpdate = ch.Last.AsTime()
			}
			result.Query = ch.Terms
		}
	case TypeTgbc:
		var ch *tgbot.Channel
		ch, err = h.svcTgBot.ReadChannel(ctx, addr)
		if err == nil {
			result.Addr = addr
			if ch.LastUpdate != nil {
				result.LastUpdate = ch.LastUpdate.AsTime()
			}
			result.Push = true
			result.Accepted = true
		}
	default:
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid source type: %s", typ))
		return
	}
	// enrich, TODO better to use GraphQL for this
	groupId := ctx.GetHeader(model.KeyGroupId)
	var limit model.Limit
	ownerId := result.UserId
	if ownerId == "" {
		ownerId = result.Addr
	}
	limit, err = h.svcLimits.Get(ctx, groupId, ownerId, model.SubjectPublishEvents)
	var usage model.Usage
	if err == nil {
		ownerId = limit.UserId
		result.Usage.Limit = limit.Count
		err = h.svcPermits.GetUsage(ctx, groupId, ownerId, model.SubjectPublishEvents, &usage)
	}
	if err == nil {
		result.Usage.Count = usage.Count
		result.Usage.Total = usage.CountTotal
	}
	if err == nil {
		switch result.UserId {
		case "":
			result.Usage.Type = UsageTypeShared
		case ctx.GetHeader(model.KeyUserId):
			// own source, leave user id set to show the delete button in UI
			result.Usage.Type = UsageTypePrivate
		default:
			// do not expose someone else's source owner user to public
			result.UserId = ""
			result.Usage.Type = UsageTypePrivate
		}
	}
	//
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, &result)
	case status.Code(err) == codes.NotFound:
		ctx.String(http.StatusNotFound, err.Error())
		return
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	return
}

func (h handler) Delete(ctx *gin.Context) {
	_, groupId, userId := grpc.AuthRequestContext(ctx)
	addrEnc := ctx.GetHeader(keySrcAddr)
	if addrEnc == "" {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("missing header: %s", keySrcAddr))
		return
	}
	addr, err := url.QueryUnescape(addrEnc)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid encoded source address: %s", addrEnc))
		return
	}
	addr = escapeNonAsciiChars(addr)
	typ := ctx.Param("type")
	switch typ {
	case TypeApub:
		err = h.svcAp.Delete(ctx, addr, groupId, userId)
	case TypeFeed:
		err = h.svcFeeds.Delete(ctx, addr, groupId, userId)
	case TypeSite:
		err = h.svcSites.Delete(ctx, addr, groupId, userId)
	case TypeTgCh:
		var ch *telegram.Channel
		ch, err = h.svcTg.Read(ctx, addr)
		if err == nil {
			if ch.GroupId == groupId && ch.UserId == userId {
				err = h.svcTg.Delete(ctx, addr)
			} else {
				ctx.String(http.StatusForbidden, "")
				return
			}
		}
	default:
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid source type: %s", typ))
		return
	}
	switch {
	case err == nil:
		ctx.String(http.StatusOK, "")
	case status.Code(err) == codes.NotFound:
		ctx.String(http.StatusNotFound, err.Error())
		return
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	return
}

func (h handler) List(ctx *gin.Context) {
	_, groupId, userId := grpc.AuthRequestContext(ctx)
	limitStr := ctx.DefaultQuery("limit", strconv.Itoa(pageLimitDefault))
	limit, err := strconv.ParseUint(limitStr, 10, 32)
	if err != nil || limit == 0 {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid limit query param: %s", limitStr))
		return
	}
	ownStr := ctx.DefaultQuery("own", "false")
	own, err := strconv.ParseBool(ownStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid own query param: %s", ownStr))
		return
	}
	var order model.Order
	orderStr := ctx.DefaultQuery("order", "ASC")
	switch orderStr {
	case "ASC":
		order = model.OrderAsc
	case "DESC":
		order = model.OrderDesc
	default:
		ctx.String(http.StatusBadRequest, "unrecognized order: %s", orderStr)
		return
	}
	pattern := ctx.DefaultQuery("filter", "")
	subId := ctx.DefaultQuery("subId", "")
	cursor := ctx.GetHeader(keySrcAddr)
	typ := ctx.Param("type")
	var page []string
	switch typ {
	case TypeApub:
		filter := &activitypub.Filter{
			Pattern: pattern,
			SubId:   subId,
		}
		if own {
			filter.GroupId = groupId
			filter.UserId = userId
		}
		page, err = h.svcAp.List(ctx, filter, uint32(limit), cursor, order)
	case TypeFeed:
		filter := &feeds.Filter{
			Pattern: pattern,
			SubId:   subId,
		}
		if own {
			filter.GroupId = groupId
			filter.UserId = userId
		}
		page, err = h.svcFeeds.ListUrls(ctx, filter, uint32(limit), cursor, order)
	case TypeSite:
		filter := &sites.Filter{
			Pattern: pattern,
		}
		if own {
			filter.GroupId = groupId
			filter.UserId = userId
		}
		page, err = h.svcSites.List(ctx, filter, uint32(limit), cursor, order)
	case TypeTgCh:
		filter := &telegram.Filter{
			Pattern: pattern,
			SubId:   subId,
		}
		if own {
			filter.GroupId = groupId
			filter.UserId = userId
		}
		page, err = h.svcTg.List(ctx, filter, uint32(limit), cursor, order)
	case TypeTgbc:
		filter := &tgbot.Filter{
			Pattern: pattern,
		}
		page, err = h.svcTgBot.ListChannels(ctx, filter, uint32(limit), cursor, order)
	default:
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid source type: %s", typ))
		return
	}
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, page)
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	return
}
