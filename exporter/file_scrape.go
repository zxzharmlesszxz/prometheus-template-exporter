package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Uint64Counter interface {
	Add(uint64) uint64
	Load() uint64
}

type FileScrapeMetrics struct {
	Path                    string
	MTimeDesc               *prometheus.Desc
	ReadErrorsTotalDesc     *prometheus.Desc
	ParseErrorsTotalDesc    *prometheus.Desc
	ScrapeDurationDesc      *prometheus.Desc
	ReadErrorsTotal         Uint64Counter
	ParseErrorsTotal        Uint64Counter
	Now                     func() time.Time
	FileModificationSeconds func(string) float64
}

func (m FileScrapeMetrics) Describe(ch chan<- *prometheus.Desc) {
	if m.MTimeDesc != nil {
		ch <- m.MTimeDesc
	}
	if m.ReadErrorsTotalDesc != nil {
		ch <- m.ReadErrorsTotalDesc
	}
	if m.ParseErrorsTotalDesc != nil {
		ch <- m.ParseErrorsTotalDesc
	}
	if m.ScrapeDurationDesc != nil {
		ch <- m.ScrapeDurationDesc
	}
}

func (m FileScrapeMetrics) Begin(ch chan<- prometheus.Metric) func() {
	start := m.now()
	if m.MTimeDesc != nil {
		ch <- prometheus.MustNewConstMetric(m.MTimeDesc, prometheus.GaugeValue, m.fileModificationSeconds(m.Path))
	}

	return func() {
		if m.ScrapeDurationDesc != nil {
			ch <- prometheus.MustNewConstMetric(m.ScrapeDurationDesc, prometheus.GaugeValue, m.since(start).Seconds())
		}
		if m.ReadErrorsTotalDesc != nil && m.ReadErrorsTotal != nil {
			ch <- prometheus.MustNewConstMetric(m.ReadErrorsTotalDesc, prometheus.CounterValue, float64(m.ReadErrorsTotal.Load()))
		}
		if m.ParseErrorsTotalDesc != nil && m.ParseErrorsTotal != nil {
			ch <- prometheus.MustNewConstMetric(m.ParseErrorsTotalDesc, prometheus.CounterValue, float64(m.ParseErrorsTotal.Load()))
		}
	}
}

func (m FileScrapeMetrics) AddReadError() {
	if m.ReadErrorsTotal != nil {
		m.ReadErrorsTotal.Add(1)
	}
}

func (m FileScrapeMetrics) AddParseError() {
	if m.ParseErrorsTotal != nil {
		m.ParseErrorsTotal.Add(1)
	}
}

func (m FileScrapeMetrics) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func (m FileScrapeMetrics) since(start time.Time) time.Duration {
	if m.Now != nil {
		return m.Now().Sub(start)
	}
	return time.Since(start)
}

func (m FileScrapeMetrics) fileModificationSeconds(path string) float64 {
	if m.FileModificationSeconds != nil {
		return m.FileModificationSeconds(path)
	}
	return FileMTimeSeconds(path)
}
