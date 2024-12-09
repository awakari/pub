package limits

import (
	"context"
	"github.com/awakari/pub/model"
	"time"
)

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Get(ctx context.Context, groupId, userId string, subj model.Subject) (l model.Limit, err error) {
	switch groupId {
	case "fail":
		err = ErrInternal
	case "missing":
	case "internal":
		l.UserId = userId
		switch userId {
		case "missing":
			l.Count = 1
		default:
			l.Count = 2
			l.Expires = time.Date(2023, 10, 1, 18, 16, 25, 0, time.UTC)
		}
	default:
		l.Count = 3
	}
	return
}
