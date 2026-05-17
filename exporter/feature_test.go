package exporter

import (
	"errors"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectorFeatureRegisterFlags(t *testing.T) {
	t.Parallel()

	called := false
	feature := CollectorFeature{
		RegisterFlagsFunc: func(app *kingpin.Application) {
			called = true
			app.Flag("demo.value", "Demo value").Default("default").String()
		},
	}

	app := kingpin.New("test", "test")
	feature.RegisterFlags(app)

	if !called {
		t.Fatal("RegisterFlagsFunc was not called")
	}
	if _, err := app.Parse([]string{"--demo.value=custom"}); err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
}

func TestCollectorFeatureRegisterCollectorsUsesCustomFunc(t *testing.T) {
	t.Parallel()

	customCalled := false
	collectorsCalled := false
	feature := CollectorFeature{
		CollectorsFunc: func(ctx FeatureContext) ([]prometheus.Collector, error) {
			collectorsCalled = true
			return nil, nil
		},
		RegisterCollectorsFunc: func(ctx FeatureContext, registry *prometheus.Registry) error {
			customCalled = true
			if ctx.Namespace != "custom_exporter" {
				t.Fatalf("FeatureContext.Namespace = %q, want %q", ctx.Namespace, "custom_exporter")
			}
			return registry.Register(newConstCollector("custom_feature_value", "Custom feature value", 3))
		},
	}

	registry := prometheus.NewRegistry()
	if err := feature.RegisterCollectors(FeatureContext{Namespace: "custom_exporter"}, registry); err != nil {
		t.Fatalf("RegisterCollectors() error = %v, want nil", err)
	}
	if !customCalled {
		t.Fatal("RegisterCollectorsFunc was not called")
	}
	if collectorsCalled {
		t.Fatal("CollectorsFunc was called despite RegisterCollectorsFunc being set")
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}
	if !hasMetricFamily(families, "custom_feature_value") {
		t.Fatal("Gather() missing custom_feature_value")
	}
}

func TestCollectorFeatureRegisterCollectorsReturnsCollectorsFuncError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("collectors failed")
	feature := CollectorFeature{
		CollectorsFunc: func(ctx FeatureContext) ([]prometheus.Collector, error) {
			return nil, wantErr
		},
	}

	err := feature.RegisterCollectors(FeatureContext{}, prometheus.NewRegistry())
	if !errors.Is(err, wantErr) {
		t.Fatalf("RegisterCollectors() error = %v, want %v", err, wantErr)
	}
}

func TestCollectorFeatureRegisterCollectorsNoopWithoutFuncs(t *testing.T) {
	t.Parallel()

	feature := CollectorFeature{}
	if err := feature.RegisterCollectors(FeatureContext{}, prometheus.NewRegistry()); err != nil {
		t.Fatalf("RegisterCollectors() error = %v, want nil", err)
	}
}

func TestCollectorFeatureRuntimeConfig(t *testing.T) {
	t.Parallel()

	if got := (CollectorFeature{}).RuntimeConfig(); got != nil {
		t.Fatalf("RuntimeConfig() = %v, want nil", got)
	}

	feature := CollectorFeature{
		RuntimeConfigFunc: func() []any {
			return []any{"demo", "enabled"}
		},
	}
	got := feature.RuntimeConfig()
	if len(got) != 2 || got[0] != "demo" || got[1] != "enabled" {
		t.Fatalf("RuntimeConfig() = %v, want [demo enabled]", got)
	}
}

func TestRegisterCollectorsReturnsDuplicateCollectorError(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collector := newConstCollector("duplicate_feature_value", "Duplicate feature value", 1)
	if err := RegisterCollectors(registry, collector); err != nil {
		t.Fatalf("RegisterCollectors() initial error = %v, want nil", err)
	}

	err := RegisterCollectors(registry, collector)
	if err == nil {
		t.Fatal("RegisterCollectors() error = nil, want duplicate collector error")
	}
}
