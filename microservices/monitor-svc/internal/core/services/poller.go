package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/monitor-svc/internal/core/ports"
)

// PollerEngine periodically schedules health checks across all active monitoring targets.
type PollerEngine struct {
	service  ports.MonitorService
	interval time.Duration
	workers  int
	logger   *slog.Logger
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewPollerEngine creates a poller engine with concurrent workers.
func NewPollerEngine(service ports.MonitorService, interval time.Duration, workers int, logger *slog.Logger) *PollerEngine {
	if workers <= 0 {
		workers = 5
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &PollerEngine{
		service:  service,
		interval: interval,
		workers:  workers,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Start launches the background polling ticker and worker pool.
func (p *PollerEngine) Start(ctx context.Context) {
	p.logger.Info("starting health check poller engine",
		slog.Duration("interval", p.interval),
		slog.Int("workers", p.workers),
	)

	ticker := time.NewTicker(p.interval)
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		defer ticker.Stop()

		// Initial sweep immediately on startup
		p.sweep(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopChan:
				return
			case <-ticker.C:
				p.sweep(ctx)
			}
		}
	}()
}

func (p *PollerEngine) sweep(ctx context.Context) {
	targets, err := p.service.ListTargets(ctx)
	if err != nil {
		p.logger.Error("failed listing targets for polling sweep", slog.String("error", err.Error()))
		return
	}

	if len(targets) == 0 {
		p.logger.Debug("no active targets to poll")
		return
	}

	jobs := make(chan domain.MonitoringTarget, len(targets))
	var workerWg sync.WaitGroup

	// Launch worker pool
	for w := 0; w < p.workers; w++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for target := range jobs {
				probeCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
				if err := p.service.CheckTarget(probeCtx, &target); err != nil {
					p.logger.Error("target probe error",
						slog.String("target_id", target.ID),
						slog.String("error", err.Error()),
					)
				}
				cancel()
			}
		}()
	}

	// Dispatch targets to workers
	for _, target := range targets {
		jobs <- target
	}
	close(jobs)

	workerWg.Wait()
}

// Stop signals the poller engine to stop and waits for in-flight probes to finish.
func (p *PollerEngine) Stop() {
	close(p.stopChan)
	p.wg.Wait()
	p.logger.Info("poller engine stopped cleanly")
}
