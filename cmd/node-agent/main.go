package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/artarts36/go-entrypoint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swarm-deploy/node-agent/internal/dockerapi"
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

	dockerClient, err := dockerapi.NewFromEnv()
	if err != nil {
		slog.Error("failed to initialize docker api client", slog.Any("err", err))
		os.Exit(1)
	}
	defer closeDockerClient(dockerClient)

	prometheus.MustRegister(metrics.NewVolumeUsageCollector(metricsNamespace, dockerClient))

	metricsMux := http.NewServeMux()
	metricsMux.Handle(metricsPath, promhttp.Handler())
	metricsServer := &http.Server{
		Addr:              metricsAddress,
		Handler:           metricsMux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	err = entrypoint.Run([]entrypoint.Entrypoint{
		entrypoint.HTTPServer("metrics-server", metricsServer),
	})
	if err != nil {
		slog.Error("[main] failed to run entrypoints", slog.Any("err", err))
	}
}

func closeDockerClient(dockerClient *dockerapi.Client) {
	err := dockerClient.Close()
	if err != nil {
		slog.Error("[main] failed to close docker client", slog.Any("err", err))
	}
}
