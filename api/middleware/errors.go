package middleware

import (
	"context"
	"net/http"

	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/sirupsen/logrus"
)

type fields interface{ Fields() map[string]interface{} }

func Fields(err error) (map[string]interface{}, bool) {
	if fe, ok := err.(fields); ok {
		return fe.Fields(), true
	}
	return nil, false
}

type response interface{ Response() (interface{}, int) }

func Response(err error) (interface{}, int, bool) {
	if re, ok := err.(response); ok {
		body, code := re.Response()
		return body, code, true
	}
	return nil, 0, false
}

func Errors(log logrus.FieldLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			err := handler(ctx, w, r)
			if err == nil {
				return nil
			}

			fields := map[string]interface{}{
				"req_id":  ContextRequestID(ctx),
				"message": err,
			}
			if f, ok := Fields(err); ok {
				for k, v := range f {
					fields[k] = v
				}
			}

			log.WithFields(logrus.Fields(fields)).Error("ERROR")

			if body, code, ok := Response(err); ok {
				return web.Respond(ctx, w, body, code)
			}

			er := struct {
				Error string `json:"error"`
			}{
				http.StatusText(http.StatusInternalServerError),
			}
			return web.Respond(ctx, w, er, http.StatusInternalServerError)
		}
		return h
	}
	return m
}
