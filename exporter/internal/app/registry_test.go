package app

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	featurepkg "github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/feature"
)

func TestNewRegistryRegistersBaseAndFeatureCollectors(t *testing.T) {
	t.Parallel()

	feature := featurepkg.CollectorFeature{
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
	if err := featurepkg.RegisterCollectors(registry, nil, newConstCollector("template_nil_skip_value", "Nil skip value", 1)); err != nil {
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

func TestNewRegistryUsesDefaultNamespaceAndSkipsNilFeatures(t *testing.T) {
	t.Parallel()

	registry, err := NewRegistry("", nil, nil)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v, want nil", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	if !hasMetricFamily(families, "exporter_framework_build_info") {
		t.Fatal("Gather() missing exporter_framework_build_info")
	}
}

func TestNewRegistryWrapsFeatureRegistrationError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("registration failed")
	feature := featurepkg.CollectorFeature{
		Name: "broken",
		RegisterCollectorsFunc: func(ctx FeatureContext, registry *prometheus.Registry) error {
			return wantErr
		},
	}

	_, err := NewRegistry("demo_exporter", nil, feature)
	if !errors.Is(err, wantErr) {
		t.Fatalf("NewRegistry() error = %v, want wrapped %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), `register feature "broken"`) {
		t.Fatalf("NewRegistry() error = %q, want feature name context", err.Error())
	}
}

func TestFeatureNameReturnsEmptyForUnnamedFeature(t *testing.T) {
	t.Parallel()

	feature := unnamedFeature{}
	if got := featureName(feature); got != "" {
		t.Fatalf("featureName() = %q, want empty string", got)
	}
}

type unnamedFeature struct{}

func (unnamedFeature) RegisterFlags(app *kingpin.Application) {}

func (unnamedFeature) RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error {
	return nil
}

func hasMetricFamily(families []*dto.MetricFamily, name string) bool {
	for _, family := range families {
		if family.GetName() == name {
			return true
		}
	}
	return false
}
