package worker

import (
	"context"
	"example.com/pki-control-plane/internal/repository"
	"log/slog"
	"time"
)

type Worker struct {
	store    repository.Store
	logger   *slog.Logger
	interval time.Duration
}

func New(s repository.Store, l *slog.Logger) *Worker {
	return &Worker{store: s, logger: l, interval: 5 * time.Second}
}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.reconcile(ctx)
		}
	}
}
func (w *Worker) reconcile(ctx context.Context) {
	for _, c := range w.store.ListCertificates(ctx) {
		if c.Status == "published" && time.Until(c.Validity.NotAfter) < 24*time.Hour {
			w.logger.Info("certificate renewal due", "certificate_id", c.ID)
		}
	}
	for _, e := range w.store.Outbox(ctx) {
		if e.DeliveredAt == nil {
			w.logger.Info("outbox pending", "event_id", e.ID)
		}
	}
}
