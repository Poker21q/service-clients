package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"service-boilerplate-go/internal/pkg/middlewares/recovery"
	"service-boilerplate-go/internal/service"

	"service-boilerplate-go/internal/api/users_auth_post"
	"service-boilerplate-go/internal/api/users_id_referrer_post"
	"service-boilerplate-go/internal/api/users_id_status_get"
	"service-boilerplate-go/internal/api/users_id_task_complete_post"
	"service-boilerplate-go/internal/api/users_leaderboard_get"
	"service-boilerplate-go/internal/pkg/middlewares/jwtauth"
	"service-boilerplate-go/internal/storage"
	"service-boilerplate-go/pkg/config"
	"service-boilerplate-go/pkg/logger"
	"service-boilerplate-go/pkg/pgdb"

	"github.com/gorilla/mux"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := logger.New()

	appConfig, err := config.Load()
	if err != nil {
		logger.Fatal(ctx, fmt.Sprintf("failed to load config %s", err))
	}

	pgdbClient, err := pgdb.New(ctx, appConfig.Postgres())
	if err != nil {
		logger.Fatal(ctx, fmt.Sprintf("failed to initialize pgdb client %s", err))
	}
	defer pgdbClient.Close()

	storageInstance := storage.New(logger, pgdbClient)
	usersService := service.New(storageInstance, appConfig.Auth().Secret())
	httpRouter := NewRouter(logger, usersService, appConfig.Auth().Secret())

	server := NewServer(appConfig.Server(), httpRouter)

	if err := runServerWithGracefulShutdown(ctx, server, logger); err != nil {
		logger.Fatal(ctx, fmt.Sprintf("server run failed: %s", err))
	}

	logger.Info(ctx, "service shutdown completed")
}

func runServerWithGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	logger *logger.Logger,
) error {
	serverErr := make(chan error, 1)

	go func() {
		logger.Info(ctx, fmt.Sprintf("starting HTTP server on %s", server.Addr))
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	logger.Info(ctx, "service started successfully")

	select {
	case err := <-serverErr:
		logger.Error(ctx, fmt.Sprintf("server error: %s, initiating shutdown", err))
	case <-ctx.Done():
		logger.Info(ctx, fmt.Sprintf("context cancelled: %s, initiating shutdown", ctx.Err()))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, fmt.Sprintf("graceful shutdown failed: %s", err))
	} else {
		logger.Info(ctx, "HTTP server shutdown completed")
	}

	return nil
}

func NewRouter(logger *logger.Logger, usersService *service.Service, secret string) http.Handler {
	router := mux.NewRouter()
	router.Use(logger.Middleware())
	router.Use(recovery.Middleware(logger))

	router.Handle("/users/auth", users_auth_post.New(logger, usersService)).Methods(http.MethodPost)

	authenticated := router.NewRoute().Subrouter()
	authenticated.Use(jwtauth.Middleware(secret))

	authenticated.Handle("/users/{id}/status", users_id_status_get.New(logger, usersService)).Methods(http.MethodGet)
	authenticated.Handle("/users/leaderboard", users_leaderboard_get.New(logger, usersService)).Methods(http.MethodGet)
	authenticated.Handle("/users/{id}/task/complete", users_id_task_complete_post.New(logger, usersService)).Methods(http.MethodPost)
	authenticated.Handle("/users/{id}/referrer", users_id_referrer_post.New(logger, usersService)).Methods(http.MethodPost)

	return router
}

// TODO: txManager

func NewServer(config config.Server, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.Host(), config.Port()),
		Handler: handler,
	}
}
