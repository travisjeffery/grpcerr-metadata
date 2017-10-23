package grpcerr

import (
	"context"
	"strconv"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type coder interface {
	Code() int
}

type withCode struct {
	error
	code int
}

func (w *withCode) Code() int { return w.code }

func wrap(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withCode{error: err, code: code}
}

const (
	codeKey = "grpcerr_code"
)

// ClientMiddleware is used on the client-requests.
func ClientMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			response, err = next(ctx, request)
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return response, err
			}
			codeVal, ok := md[codeKey]
			if !ok {
				return response, err
			}
			code, err := strconv.Atoi(codeVal[0])
			if err != nil {
				return response, err
			}
			return response, wrap(err, code)
		}
	}
}

// ServerMiddleware is used on the server-responses.
func ServerMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			response, err = next(ctx, request)
			if err == nil {
				return response, nil
			}
			type causer interface {
				Cause() error
			}
			for err != nil {
				cause, ok := err.(causer)
				if !ok {
					break
				}
				err = cause.Cause()
			}
			if errsc, ok := err.(coder); ok {
				header := metadata.Pairs(codeKey, strconv.Itoa(errsc.Code()))
				grpc.SendHeader(ctx, header)
			}
			return response, err
		}
	}
}
