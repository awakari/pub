package permits

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/pub/api/grpc/auth"
	"github.com/awakari/pub/api/grpc/subject"
	"github.com/awakari/pub/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type Service interface {

	// GetUsage returns the current group/user spent counts for the specified subject.
	// A client should specify the empty user id to get the group-level value.
	GetUsage(ctx context.Context, groupId, userId string, subj model.Subject, out *model.Usage) (err error)

	// Request attempts to acknowledge a certain requested count.
	// The returned Permit contains:
	// 	* userId which is empty for the group-level Permit
	// 	* acknowledged count
	Request(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (p model.Permit, err error)

	// Release decrements the usage counter by the specified count when a client fails to utilize all permits returned
	// by the preceding Request call. The userId argument value must be taken from the Permit returned by the
	// preceding Request call.
	Release(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (err error)
}

type service struct {
	client ServiceClient
}

var ErrInternal = errors.New("internal failure")
var ErrInvalid = errors.New("invalid")
var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("usage record not found")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) GetUsage(ctx context.Context, groupId, userId string, subj model.Subject, u *model.Usage) (err error) {
	ctxAuth := auth.SetOutgoingAuthInfo(ctx, groupId, userId)
	var resp *GetResponse
	var reqSubj subject.Subject
	reqSubj, err = subject.Encode(subj)
	resp, err = svc.client.Get(ctxAuth, &GetRequest{
		Subj: reqSubj,
	})
	if resp != nil {
		u.Count = resp.Count
		u.CountTotal = resp.CountTotal
		if resp.Since != nil {
			u.Since = resp.Since.AsTime().UTC()
		}
	}
	err = decodeError(err)
	return
}

func (svc service) Request(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (u model.Permit, err error) {
	var resp *AllocateResponse
	var reqSubj subject.Subject
	reqSubj, err = subject.Encode(subj)
	if err == nil {
		resp, err = svc.client.Allocate(ctx, &AllocateRequest{
			GroupId: groupId,
			UserId:  userId,
			Count:   count,
			Subj:    reqSubj,
		})
	}
	if resp != nil {
		u.Count = resp.Count
		u.UserId = resp.UserId
		u.JustExhausted = resp.JustExhausted
	}
	err = decodeError(err)
	return
}

func (svc service) Release(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (err error) {
	var reqSubj subject.Subject
	reqSubj, err = subject.Encode(subj)
	if err == nil {
		_, err = svc.client.Release(ctx, &ReleaseRequest{
			GroupId: groupId,
			UserId:  userId,
			Count:   count,
			Subj:    reqSubj,
		})
	}
	err = decodeError(err)
	return
}

func decodeError(src error) (dst error) {
	switch {
	case src == io.EOF:
		dst = src // return as it is
	case status.Code(src) == codes.OK:
		dst = nil
	case status.Code(src) == codes.InvalidArgument:
		dst = fmt.Errorf("%w: %s", ErrInvalid, src)
	case status.Code(src) == codes.NotFound:
		dst = fmt.Errorf("%w: %s", ErrNotFound, src)
	case status.Code(src) == codes.Unauthenticated:
		dst = fmt.Errorf("%w: %s", ErrForbidden, src)
	default:
		dst = fmt.Errorf("%w: %s", ErrInternal, src)
	}
	return
}
