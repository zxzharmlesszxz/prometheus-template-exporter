package exporter

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zxzharmlesszxz/prometheus-template-exporter/exporter/exportertest"
)

func TestFileScrapeMetricsDescribeAndCollect(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	readErrors := atomic.Uint64{}
	parseErrors := atomic.Uint64{}
	mtimeDesc := prometheus.NewDesc("file_mtime_seconds", "File mtime.", nil, nil)
	readErrorsDesc := prometheus.NewDesc("file_read_errors_total", "File read errors.", nil, nil)
	parseErrorsDesc := prometheus.NewDesc("file_parse_errors_total", "File parse errors.", nil, nil)
	durationDesc := prometheus.NewDesc("file_scrape_duration_seconds", "File scrape duration.", nil, nil)

	collector := fileScrapeTestCollector{
		metrics: FileScrapeMetrics{
			Path:                 "/tmp/input",
			MTimeDesc:            mtimeDesc,
			ReadErrorsTotalDesc:  readErrorsDesc,
			ParseErrorsTotalDesc: parseErrorsDesc,
			ScrapeDurationDesc:   durationDesc,
			ReadErrorsTotal:      &readErrors,
			ParseErrorsTotal:     &parseErrors,
			Now: func() time.Time {
				return now
			},
			FileModificationSeconds: func(path string) float64 {
				if path != "/tmp/input" {
					t.Fatalf("path = %q, want /tmp/input", path)
				}
				return 123
			},
		},
		collect: func(metrics FileScrapeMetrics) {
			now = now.Add(250 * time.Millisecond)
			metrics.AddReadError()
			metrics.AddParseError()
		},
	}

	families := exportertest.RegisterAndGather(t, collector)
	exportertest.AssertMetricValue(t, families, "file_mtime_seconds", nil, 123)
	exportertest.AssertMetricValue(t, families, "file_read_errors_total", nil, 1)
	exportertest.AssertMetricValue(t, families, "file_parse_errors_total", nil, 1)
	exportertest.AssertMetricValue(t, families, "file_scrape_duration_seconds", nil, 0.25)
}

func TestFileScrapeMetricsAllowsOptionalCounters(t *testing.T) {
	t.Parallel()

	mtimeDesc := prometheus.NewDesc("optional_file_mtime_seconds", "File mtime.", nil, nil)
	collector := fileScrapeTestCollector{
		metrics: FileScrapeMetrics{
			Path:      "/tmp/input",
			MTimeDesc: mtimeDesc,
			FileModificationSeconds: func(string) float64 {
				return 456
			},
		},
		collect: func(metrics FileScrapeMetrics) {
			metrics.AddReadError()
			metrics.AddParseError()
		},
	}

	families := exportertest.RegisterAndGather(t, collector)
	exportertest.AssertMetricValue(t, families, "optional_file_mtime_seconds", nil, 456)
}

type fileScrapeTestCollector struct {
	metrics FileScrapeMetrics
	collect func(FileScrapeMetrics)
}

func (c fileScrapeTestCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.Describe(ch)
}

func (c fileScrapeTestCollector) Collect(ch chan<- prometheus.Metric) {
	finish := c.metrics.Begin(ch)
	defer finish()
	if c.collect != nil {
		c.collect(c.metrics)
	}
}
