package subject

import (
	"fmt"
	"github.com/awakari/pub/model"
)

func Encode(src model.Subject) (dst Subject, err error) {
	switch src {
	case model.SubjectInterests:
		dst = Subject_Interests
	case model.SubjectPublishHourly:
		dst = Subject_PublishHourly
	case model.SubjectPublishDaily:
		dst = Subject_PublishDaily
	default:
		err = fmt.Errorf("invalid subject: %s", src)
	}
	return
}
