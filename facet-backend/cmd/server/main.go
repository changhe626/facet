package main

import (
	"context"
	"facet-backend/internal/server"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		panic("failed to create server: " + err.Error())
	}

	srv.Start()

	httpServer := &http.Server{
		Addr:    ":" + srv.Cfg().Server.Port,
		Handler: srv.Engine(),
	}

	go func() {
		slog.Info("starting server", "port", srv.Cfg().Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited")
}
