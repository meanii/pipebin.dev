package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

		logger, err := cfg.Build(
			zap.WithCaller(false),
			zap.AddCallerSkip(0),
		)
		if err != nil {
			panic(err)
		}
		defer logger.Sync()

		reqLogger := logger.With()

		ctx := context.WithValue(r.Context(), "logger", reqLogger)
		r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
