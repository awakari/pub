package subject

import (
	"fmt"
	"github.com/awakari/pub/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Decode(src Subject) (dst model.Subject, err error) {
	switch src {
	case Subject_Interests:
		dst = model.SubjectInterests
	case Subject_PublishEvents:
		dst = model.SubjectPublishEvents
	default:
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("invalid subject: %s", src))
	}
	return
}

func Encode(src model.Subject) (dst Subject, err error) {
	switch src {
	case model.SubjectInterests:
		dst = Subject_Interests
	case model.SubjectPublishEvents:
		dst = Subject_PublishEvents
	default:
		err = fmt.Errorf(fmt.Sprintf("invalid subject: %s", src))
	}
	return
}
