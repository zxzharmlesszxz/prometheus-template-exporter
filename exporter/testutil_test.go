package exporter

import "github.com/prometheus/client_golang/prometheus"

type constCollector struct {
	desc  *prometheus.Desc
	value float64
}

func newConstCollector(name string, help string, value float64) prometheus.Collector {
	return constCollector{
		desc:  prometheus.NewDesc(name, help, nil, nil),
		value: value,
	}
}

func (c constCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c constCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, c.value)
}
