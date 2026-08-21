package main

import (
	"context"
	"example.com/pki-control-plane/internal/repository"
	"example.com/pki-control-plane/internal/worker"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	w := worker.New(repository.NewMemoryStore(), l)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	w.Run(ctx)
	time.Sleep(10 * time.Millisecond)
}
