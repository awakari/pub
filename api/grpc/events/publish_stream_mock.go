package events

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
)

type publishStreamMock struct {
	lastReq *PublishRequest
	lastErr error
}

func newPublishStreamMock() Service_PublishClient {
	return &publishStreamMock{}
}

func (sm *publishStreamMock) Send(req *PublishRequest) (err error) {
	switch sm.lastErr {
	case nil:
		switch req.Topic {
		case "recv_fail":
			sm.lastErr = status.Error(codes.Internal, "send failure")
		case "send_eof":
			sm.lastErr = io.EOF
			err = io.EOF
		case "recv_eof":
			sm.lastErr = io.EOF
		case "missing":
			sm.lastErr = status.Error(codes.NotFound, "queue missing")
		default:
			sm.lastReq = req
		}
	default:
		err = io.EOF
	}
	return
}

func (sm *publishStreamMock) Recv() (resp *PublishResponse, err error) {
	resp = &PublishResponse{}
	switch sm.lastErr {
	case nil:
		resp.AckCount = uint32(len(sm.lastReq.Evts))
		if resp.AckCount > 2 {
			resp.AckCount = 2
		}
	default:
		err = sm.lastErr
	}
	return
}

func (sm *publishStreamMock) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (sm *publishStreamMock) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (sm *publishStreamMock) CloseSend() error {
	return nil
}

func (sm *publishStreamMock) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (sm *publishStreamMock) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (sm *publishStreamMock) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}
