package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/sirupsen/logrus"
	"github.com/zenazn/goji/web/mutil"
)

func Logger(log logrus.FieldLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			log := log

			if rid := ContextRequestID(ctx); rid != "" {
				log = log.WithField("req_id", rid)
			}

			log = log.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteaddr": r.RemoteAddr,
			})

			log.Info("started")
			startTime := time.Now().UTC()

			lw := mutil.WrapWriter(w)
			err := handler(ctx, lw, r)

			log = log.WithFields(logrus.Fields{
				"statuscode": lw.Status(),
				"bytes":      lw.BytesWritten(),
				"since":      time.Since(startTime).Nanoseconds(),
			})
			log.Info("completed")
			return err
		}
		return h
	}
	return m
}
