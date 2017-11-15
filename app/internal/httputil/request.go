package httputil

import (
	"context"
	"github.com/rs/xid"
	"net/http"
)

type requestId string

const (
	reqId           requestId = "request_id"
	HeaderRequestId           = "X-Request-Id"
)

func AddRequestId(ctx context.Context, r *http.Request) context.Context {
	var id string
	if r.Header.Get(HeaderRequestId) != "" {
		id = r.Header.Get(HeaderRequestId)
	} else {
		id = xid.New().String()
	}
	return context.WithValue(ctx, reqId, id)
}

func AddCustomRequestId(ctx context.Context, value string) context.Context {
	if len(value) == 0 {
		value = xid.New().String()
	}

	return context.WithValue(ctx, reqId, value)
}

func GetRequestId(ctx context.Context) string {
	value := ctx.Value(reqId)

	switch value.(type) {
	case string:
		return value.(string)

	default:
		return ""
	}
}
