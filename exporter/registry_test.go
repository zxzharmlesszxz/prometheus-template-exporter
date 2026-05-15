package exporter

import (
	"io"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestNewRegistryRegistersBaseAndFeatureCollectors(t *testing.T) {
	t.Parallel()

	feature := CollectorFeature{
		Name: "demo",
		CollectorsFunc: func(ctx FeatureContext) ([]prometheus.Collector, error) {
			if ctx.Namespace != "demo_exporter" {
				t.Fatalf("FeatureContext.Namespace = %q, want %q", ctx.Namespace, "demo_exporter")
			}
			if ctx.Logger == nil {
				t.Fatal("FeatureContext.Logger = nil, want logger")
			}
			return []prometheus.Collector{
				newConstCollector("demo_feature_value", "Demo feature value", 1),
			}, nil
		},
	}

	registry, err := NewRegistry("demo_exporter", slog.New(slog.NewTextHandler(io.Discard, nil)), feature)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v, want nil", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	if !hasMetricFamily(families, "demo_exporter_build_info") {
		t.Fatal("Gather() missing demo_exporter_build_info")
	}
	if !hasMetricFamily(families, "demo_feature_value") {
		t.Fatal("Gather() missing demo_feature_value")
	}
}

func TestRegisterCollectorsSkipsNilCollectors(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	if err := RegisterCollectors(registry, nil, newConstCollector("template_nil_skip_value", "Nil skip value", 1)); err != nil {
		t.Fatalf("RegisterCollectors() error = %v, want nil", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	if !hasMetricFamily(families, "template_nil_skip_value") {
		t.Fatal("Gather() missing template_nil_skip_value")
	}
}

func hasMetricFamily(families []*dto.MetricFamily, name string) bool {
	for _, family := range families {
		if family.GetName() == name {
			return true
		}
	}
	return false
}
