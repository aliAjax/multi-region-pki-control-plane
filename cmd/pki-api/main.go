package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/pki-control-plane/internal/config"
	"example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/issuance"
	"example.com/pki-control-plane/internal/repository"
	"example.com/pki-control-plane/internal/revocation"
	"example.com/pki-control-plane/internal/transport/httpapi"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := repository.NewMemoryStore()
	signer := crypto.NewSignerFactory(cfg.DevelopmentHSM)
	issuer := issuance.NewService(store, signer)
	revoker := revocation.NewService(store)
	h := httpapi.NewServer(cfg, logger, store, issuer, revoker, signer)
	srv := &http.Server{Addr: cfg.ListenAddr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("pki api started", "addr", cfg.ListenAddr, "unsafe_dev_hsm", cfg.DevelopmentHSM)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "error", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
