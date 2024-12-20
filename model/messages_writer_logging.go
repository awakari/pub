package model

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
)

type messagesWriterLogging struct {
	w    MessagesWriter
	log  *slog.Logger
	name string
}

func NewMessagesWriterLogging(w MessagesWriter, log *slog.Logger, name string) MessagesWriter {
	return messagesWriterLogging{
		w:    w,
		log:  log,
		name: name,
	}
}

func (lw messagesWriterLogging) Close() (err error) {
	err = lw.w.Close()
	ll := lw.logLevel(err)
	lw.log.Log(context.TODO(), ll, fmt.Sprintf("messages.writer(%s).Close(): %s", lw.name, err))
	return
}

func (lw messagesWriterLogging) Write(ctx context.Context, msgs []*pb.CloudEvent) (ackCount uint32, err error) {
	ackCount, err = lw.w.Write(ctx, msgs)
	ll := lw.logLevel(err)
	if err != nil {
		var ids []string
		for _, msg := range msgs {
			ids = append(ids, msg.Id)
		}
		err = fmt.Errorf("%w\nmessage ids: %+v", err, ids)
	}
	lw.log.Log(ctx, ll, fmt.Sprintf("msgWriter(%s).Write(count=%d): ack=%d, err=%s", lw.name, len(msgs), ackCount, err))
	return
}

func (lw messagesWriterLogging) logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelError
	}
	return
}
