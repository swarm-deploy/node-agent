package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	collectTimeout    = 15 * time.Second
	megabyteDelimiter = 1024 * 1024
)

// UsageCollector exports docker volume usage metrics for Prometheus.
type UsageCollector struct {
	docker *client.Client

	usageMegabytes *prometheus.Desc
}

// NewVolumeUsageCollector creates Prometheus collector for docker volumes.
func NewVolumeUsageCollector(namespace string, dockerClient *client.Client) *UsageCollector {
	return &UsageCollector{
		docker: dockerClient,
		usageMegabytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "usage_megabytes"),
			"Docker volume usage in megabytes.",
			[]string{"volume", "driver", "scope"},
			nil,
		),
	}
}

// Describe sends metric descriptors to Prometheus.
func (c *UsageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.usageMegabytes
}

// Collect sends fresh volume usage metrics to Prometheus.
func (c *UsageCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), collectTimeout)
	defer cancel()

	diskUsage, err := c.docker.DiskUsage(ctx, types.DiskUsageOptions{Types: []types.DiskUsageObject{types.VolumeObject}})
	if err != nil {
		slog.ErrorContext(ctx, "failed to collect volume usage metrics", slog.Any("err", err))
		ch <- prometheus.NewInvalidMetric(c.usageMegabytes, err)
		return
	}

	for _, vol := range diskUsage.Volumes {
		if vol == nil || vol.UsageData == nil || isAnonymousVolume(vol) {
			continue
		}

		labels := []string{vol.Name, vol.Driver, vol.Scope}
		size := float64(vol.UsageData.Size / megabyteDelimiter)
		ch <- prometheus.MustNewConstMetric(c.usageMegabytes, prometheus.GaugeValue, size, labels...)
	}
}

func isAnonymousVolume(vol *volume.Volume) bool {
	_, hasLabel := vol.Labels["com.docker.volume.anonymous"]
	return hasLabel
}
