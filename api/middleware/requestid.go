package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/irsalhamdi/e-commerce-video/api/web"

	"context"
)

const (
	RequestIDHeader = "X-Request-Id"

	DefaultRequestIDLengthLimit = 128
)

type reqIDKeyCtx int

const reqIDKey reqIDKeyCtx = 1

var reqID int64

var reqPrefix string

func init() {
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	reqPrefix = string(b64[0:10])
}

func RequestID() web.Middleware {
	lengthLimit := DefaultRequestIDLengthLimit
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			id := r.Header.Get(RequestIDHeader)
			if id == "" {
				id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
			} else if lengthLimit >= 0 && len(id) > lengthLimit {
				id = id[:lengthLimit]
			}
			ctx = context.WithValue(ctx, reqIDKey, id)

			return handler(ctx, w, r)
		}
		return h
	}
	return m
}

func ContextRequestID(ctx context.Context) (reqID string) {
	id := ctx.Value(reqIDKey)
	if id != nil {
		reqID = id.(string)
	}
	return
}
