package auth

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	Authenticate(ctx context.Context, userId, token string) (err error)
}

type service struct {
	client ServiceClient
}

var ErrInvalidToken = errors.New("invalid token")
var ErrInvalidUserId = errors.New("invalid user id")
var ErrInternal = errors.New("internal failure")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Authenticate(ctx context.Context, userId, token string) (err error) {
	_, err = svc.client.Authenticate(ctx, &AuthenticateRequest{
		UserId: userId,
		Token:  token,
	})
	switch status.Code(err) {
	case codes.OK:
		err = nil
	case codes.Unauthenticated:
		err = fmt.Errorf("%w: user=%s, token=%s", ErrInvalidToken, userId, token)
	case codes.InvalidArgument:
		err = fmt.Errorf("%w: %s", ErrInvalidUserId, userId)
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return err
}
