package grpcerr

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type response struct {
	v string
}

func TestMiddleware(t *testing.T) {
	ctx := context.Background()
	smw := ServerMiddleware()
	sep := smw(func(ctx context.Context, request interface{}) (interface{}, error) {
		return response{"the response"}, wrap(errors.New("not found"), http.StatusNotFound)
	})
	cmw := ClientMiddleware()
	cep := cmw(sep)

	t.Run("server middleware", func(t *testing.T) {
		resp, err := sep(ctx, &http.Request{})
		if err == nil {
			t.Error("expected error; got none")
		}
		r, ok := resp.(response)
		if !ok {
			t.Error("expected response")
		}
		if r.v != "the response" {
			t.Errorf("got %s; want %s", r.v, "the response")
		}
	})

	t.Run("client middleware", func(t *testing.T) {
		resp, err := cep(ctx, &http.Request{})
		if err == nil {
			t.Error("expected error; got none")
		}
		r, ok := resp.(response)
		if !ok {
			t.Error("expected response")
		}
		if r.v != "the response" {
			t.Errorf("got %s; want %s", r.v, "the response")
		}
		esc, ok := err.(coder)
		if !ok {
			t.Error("expected statusCoder")
		}
		if esc.StatusCode() != http.StatusNotFound {
			t.Errorf("got %d; want %d", esc.StatusCode(), http.StatusNotFound)
		}
	})
}
