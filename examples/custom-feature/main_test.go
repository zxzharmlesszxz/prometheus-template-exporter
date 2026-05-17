package main

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	template "github.com/zxzharmlesszxz/prometheus-template-exporter/exporter"
)

func TestFeatureRegistersDemoCollector(t *testing.T) {
	t.Parallel()

	feature := &Feature{}
	app := kingpin.New("test", "test")
	feature.RegisterFlags(app)
	if _, err := app.Parse([]string{"--demo.target=node-a"}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	registry := prometheus.NewRegistry()
	err := feature.RegisterCollectors(template.FeatureContext{Namespace: "demo_exporter"}, registry)
	if err != nil {
		t.Fatalf("RegisterCollectors() error = %v, want nil", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	metric := findMetricFamily(families, "demo_exporter_demo_info")
	if metric == nil {
		t.Fatal("Gather() missing demo_exporter_demo_info")
	}
	if got := metric.GetMetric()[0].GetGauge().GetValue(); got != 1 {
		t.Fatalf("demo_exporter_demo_info = %v, want 1", got)
	}
	if got := metric.GetMetric()[0].GetLabel()[0].GetValue(); got != "node-a" {
		t.Fatalf("demo_exporter_demo_info target = %q, want %q", got, "node-a")
	}
}

func TestFeatureRuntimeConfigAndDefaultListenAddress(t *testing.T) {
	t.Parallel()

	feature := &Feature{}
	app := kingpin.New("test", "test")
	feature.RegisterFlags(app)
	if _, err := app.Parse([]string{"--demo.target=node-b"}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got := feature.DefaultListenAddress(); got != ":9901" {
		t.Fatalf("DefaultListenAddress() = %q, want %q", got, ":9901")
	}
	runtimeConfig := feature.RuntimeConfig()
	if len(runtimeConfig) != 2 || runtimeConfig[0] != "demo_target" || runtimeConfig[1] != "node-b" {
		t.Fatalf("RuntimeConfig() = %v, want [demo_target node-b]", runtimeConfig)
	}
}

func findMetricFamily(families []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}
