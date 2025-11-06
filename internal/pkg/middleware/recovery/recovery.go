package recovery

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"service-boilerplate-go/internal/pkg/response"
)

type Logger interface {
	Error(ctx context.Context, msg string)
}

func Middleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(r.Context(), fmt.Sprintf("panic recovered %s, %s", err, debug.Stack()))
					response.ErrorStatus(w, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
