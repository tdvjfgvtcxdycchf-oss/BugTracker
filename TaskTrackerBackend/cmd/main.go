package main

import (
	"bug_tracker/internal/sql"
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	logPath := "logs/app.log"
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		slog.Error("error", err)
		os.Exit(1)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(logger)

	slog.Info("starting bug tracker application")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := sql.CreateConnection(ctx)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	slog.Info("connected to database")

}
