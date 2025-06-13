//go:build !solution

package requestlog

import (
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	mdw := func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method
			requestID := uuid.Must(uuid.NewV4()).String()

			l.Info("request started",
				zap.String("path", path),
				zap.String("method", method),
				zap.String("request_id", requestID),
			)
			start := time.Now()
			defer func() {
				if r := recover(); r != nil {
					l.Info("request panicked",
						zap.String("path", path),
						zap.String("method", method),
						zap.String("request_id", requestID),
					)
					panic(r)
				}
			}()
			metr := httpsnoop.CaptureMetrics(next, w, r)
			duration := time.Since(start)
			l.Info("request finished",
				zap.String("path", path),
				zap.String("method", method),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.Int("status_code", metr.Code),
			)
		}
		return http.HandlerFunc(f)
	}
	return mdw
}
