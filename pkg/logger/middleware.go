package logger

import (
	"net/http"
)

func (l *Logger) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Добавляем только базовую информацию о запросе
			ctx = l.WithFields(ctx, map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})

			// Подменяем контекст у запроса
			r = r.WithContext(ctx)

			// Передаём дальше без логирования
			next.ServeHTTP(w, r)
		})
	}
}
