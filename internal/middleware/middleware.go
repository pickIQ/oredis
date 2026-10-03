package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/pickIQ/oredis/internal/response"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stacktrace := string(debug.Stack())
				slog.Error("panic recovered", "error", rec, "path", r.URL.Path, "stacktrace", stacktrace)
				response.Error(w, http.StatusInternalServerError, "something went wrong")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		defer func() {
			slog.Info("HTTP Request:", "method", r.Method, "url", r.URL.Path, "duration", time.Since(start), "code", rec.status)
		}()
		next.ServeHTTP(rec, r)

	})
}
