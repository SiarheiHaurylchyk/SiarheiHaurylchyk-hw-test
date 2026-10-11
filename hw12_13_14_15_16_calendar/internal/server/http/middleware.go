package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
		logger.Info(fmt.Sprintf("%s [%s] %s %s %s %d %d %q",
			clientIP,
			startTime.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.RequestURI(),
			r.Proto,
			recorder.statusCode,
			time.Since(startTime).Microseconds(),
			r.UserAgent(),
		))
	})
}
