package permits

import (
	"context"
	"github.com/awakari/pub/api/grpc/limits"
	"github.com/awakari/pub/model"
	"time"
)

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) GetUsage(ctx context.Context, groupId, userId string, subj model.Subject, u *model.Usage) (err error) {
	switch groupId {
	case "fail":
		err = ErrInternal
	default:
		u.Count = 1
		u.CountTotal = 2
		u.Since = time.Date(2023, 05, 07, 04, 57, 20, 0, time.UTC)
	}
	return
}

func (sm serviceMock) Request(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (p model.Permit, err error) {
	switch groupId {
	case "fail":
		err = ErrInternal
	case "limits_fail":
		err = limits.ErrInternal
	case "limit_reached":
		if userId != "external" {
			p.UserId = userId
		}
	default:
		if userId != "external" {
			p.UserId = userId
		}
		if count > 2 {
			p.Count = 2
		} else {
			p.Count = count
		}
	}
	return
}

func (sm serviceMock) Release(ctx context.Context, groupId, userId string, subj model.Subject, count uint32) (err error) {
	switch groupId {
	case "fail":
		err = ErrInternal
	}
	return
}
