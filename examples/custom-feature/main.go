package main

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"

	template "github.com/zxzharmlesszxz/prometheus-template-exporter/exporter"
)

type Feature struct {
	target *string
}

func (f *Feature) RegisterFlags(app *kingpin.Application) {
	f.target = app.Flag("demo.target", "Target name exposed by the demo collector").Default("local").String()
}

func (f *Feature) DefaultListenAddress() string {
	return ":9901"
}

func (f *Feature) RegisterCollectors(ctx template.FeatureContext, registry *prometheus.Registry) error {
	target := "local"
	if f.target != nil {
		target = *f.target
	}

	desc := prometheus.NewDesc(
		ctx.Namespace+"_demo_info",
		"Demo feature metadata.",
		[]string{"target"},
		nil,
	)
	return template.RegisterCollectors(registry, demoCollector{
		desc:   desc,
		target: target,
	})
}

func (f *Feature) RuntimeConfig() []any {
	target := "local"
	if f.target != nil {
		target = *f.target
	}
	return []any{"demo_target", target}
}

type demoCollector struct {
	desc   *prometheus.Desc
	target string
}

func (c demoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c demoCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, 1, c.target)
}

func main() {
	template.MainFromProject(&Feature{})
}
