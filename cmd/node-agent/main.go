package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/artarts36/go-entrypoint"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swarm-deploy/node-agent/internal/cleaner"
	"github.com/swarm-deploy/node-agent/internal/config"
	"github.com/swarm-deploy/node-agent/internal/metrics"
)

const (
	metricsNamespace  = "swarm_deploy_node_agent"
	metricsAddress    = ":9000"
	metricsPath       = "/metrics"
	readHeaderTimeout = 5 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("[main] read config")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to read environment variables", slog.Any("err", err))
		os.Exit(1)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		slog.Error("failed to initialize docker api client", slog.Any("err", err))
		os.Exit(1)
	}

	prometheus.MustRegister(metrics.NewVolumeUsageCollector(metricsNamespace, dockerClient))

	metricsMux := http.NewServeMux()
	metricsMux.Handle(metricsPath, promhttp.Handler())
	metricsServer := &http.Server{
		Addr:              metricsAddress,
		Handler:           metricsMux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	entrypoints := []entrypoint.Entrypoint{
		entrypoint.HTTPServer("metrics-server", metricsServer),
	}

	if cfg.Cleaner.Interval > 0 {
		clean, cerr := cleaner.New(cfg.Cleaner, dockerClient)
		if cerr != nil {
			slog.Error("failed to create cleaner", slog.Any("err", err))
			closeDockerClient(dockerClient)
			os.Exit(1)
		}

		entrypoints = append(entrypoints, entrypoint.Entrypoint{
			Name: "cleaner",
			Run: func(ctx context.Context) error {
				clean.Run(ctx)
				return nil
			},
		})
	}

	err = entrypoint.Run(entrypoints)
	if err != nil {
		slog.Error("[main] failed to run entrypoints", slog.Any("err", err))
	}
	closeDockerClient(dockerClient)
}

func closeDockerClient(dockerClient *client.Client) {
	err := dockerClient.Close()
	if err != nil {
		slog.Error("[main] failed to close docker client", slog.Any("err", err))
	}
}
