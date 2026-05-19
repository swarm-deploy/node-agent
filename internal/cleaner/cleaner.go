package cleaner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	dcleaner "github.com/artarts36/docker-cleanup/pkg/cleaner"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/swarm-deploy/node-agent/internal/config"
)

type Cleaner struct {
	cfg     config.Cleaner
	cleaner dcleaner.Cleaner
}

func New(cfg config.Cleaner, dockerClient *client.Client) (*Cleaner, error) {
	metricsCollector := dcleaner.NewPrometheusMetricsCollector("swarm_deploy_node_agent")

	cleaner := dcleaner.New(dockerClient, dcleaner.Opts{
		Containers:       true,
		Images:           true,
		MetricsCollector: metricsCollector,
	})

	if err := prometheus.Register(metricsCollector); err != nil {
		return nil, fmt.Errorf("register metrics: %w", err)
	}

	return &Cleaner{
		cfg:     cfg,
		cleaner: cleaner,
	}, nil
}

func (c *Cleaner) Run(ctx context.Context) {
	t := time.NewTicker(c.cfg.Interval)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "[cleaner] stopped")
		case <-t.C:
			err := c.cleaner.Clean(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "[cleaner] failed to clean", slog.Any("err", err))
			}
		}
	}
}
