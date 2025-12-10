package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/kanshan/ServerStatus/data-server/internal/app"
	"github.com/kanshan/ServerStatus/data-server/internal/handler"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	components, err := app.Build(*configPath)
	if err != nil {
		panic(err)
	}

	logger := components.Logger
	cfg := components.Config

	router, err := handler.BuildHTTPHandler(&handler.HandlerDependencies{
		ServerService:    components.ServerService,
		ExportService:    components.ExportService,
		AuthService:      components.AuthService,
		HealthService:    components.HealthService,
		WebSocketService: components.WebSocketService,
		Logger:           components.Logger,
	}, handler.DefaultHandlerConfig())
	if err != nil {
		logger.WithError(err).Fatal("failed to build http handler")
	}

	handlerWithMiddleware := gziphandler.GzipHandler(router)

	server := &http.Server{
		Addr:         cfg.Server.GetAddr(),
		Handler:      handlerWithMiddleware,
		ReadTimeout:  parseDuration(cfg.Server.ReadTimeout, 30*time.Second),
		WriteTimeout: parseDuration(cfg.Server.WriteTimeout, 30*time.Second),
		IdleTimeout:  parseDuration(cfg.Server.IdleTimeout, time.Minute),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Graceful shutdown failed")
		}
	}()

	logger.Infof("Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Fatal("Server exited with error")
	}
	logger.Info("Server stopped")
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseDuration(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}
