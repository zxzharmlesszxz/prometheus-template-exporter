package exporter

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func TestFacadeInjectedMetadataUsesRootPackageVariables(t *testing.T) {
	withFacadeInjectedMetadata(t, ProjectMetadata{
		ExporterName:         "prometheus-facade-exporter",
		ExporterDescription:  "Prometheus Facade Exporter",
		FeatureName:          "facade",
		MetricNamespace:      "facade_exporter",
		DefaultListenAddress: ":9123",
	})

	if InjectedExporterName() != "prometheus-facade-exporter" {
		t.Fatalf("InjectedExporterName() = %q", InjectedExporterName())
	}
	if InjectedExporterDescription() != "Prometheus Facade Exporter" {
		t.Fatalf("InjectedExporterDescription() = %q", InjectedExporterDescription())
	}
	if InjectedFeatureName() != "facade" {
		t.Fatalf("InjectedFeatureName() = %q", InjectedFeatureName())
	}
	if InjectedMetricNamespace() != "facade_exporter" {
		t.Fatalf("InjectedMetricNamespace() = %q", InjectedMetricNamespace())
	}
	if InjectedDefaultListenAddress() != ":9123" {
		t.Fatalf("InjectedDefaultListenAddress() = %q", InjectedDefaultListenAddress())
	}

	cfg := ConfigFromInjectedProject(CollectorFeature{Name: "facade"})
	if cfg.Name != "prometheus-facade-exporter" {
		t.Fatalf("ConfigFromInjectedProject().Name = %q", cfg.Name)
	}
	if cfg.Namespace != "facade_exporter" {
		t.Fatalf("ConfigFromInjectedProject().Namespace = %q", cfg.Namespace)
	}
	if cfg.DefaultListenAddress != ":9123" {
		t.Fatalf("ConfigFromInjectedProject().DefaultListenAddress = %q", cfg.DefaultListenAddress)
	}

	info := ExporterInfoFromInjectedProject(facadeInjectedSmokeFeature{})
	if info.Name != "prometheus-facade-exporter" {
		t.Fatalf("ExporterInfoFromInjectedProject().Name = %q", info.Name)
	}
	if !hasFacadeTestString(info.Smoke.WantMetrics, "facade_exporter_custom_metric 1") {
		t.Fatalf("ExporterInfoFromInjectedProject().Smoke.WantMetrics = %v", info.Smoke.WantMetrics)
	}
}

func withFacadeInjectedMetadata(t *testing.T, metadata ProjectMetadata) {
	t.Helper()

	oldExporterName := injectedExporterName
	oldExporterDescription := injectedExporterDescription
	oldFeatureName := injectedFeatureName
	oldMetricNamespace := injectedMetricNamespace
	oldListenAddress := injectedListenAddress
	t.Cleanup(func() {
		injectedExporterName = oldExporterName
		injectedExporterDescription = oldExporterDescription
		injectedFeatureName = oldFeatureName
		injectedMetricNamespace = oldMetricNamespace
		injectedListenAddress = oldListenAddress
	})

	injectedExporterName = metadata.ExporterName
	injectedExporterDescription = metadata.ExporterDescription
	injectedFeatureName = metadata.FeatureName
	injectedMetricNamespace = metadata.MetricNamespace
	injectedListenAddress = metadata.DefaultListenAddress
}

type facadeInjectedSmokeFeature struct{}

func (facadeInjectedSmokeFeature) RegisterFlags(*kingpin.Application) {}

func (facadeInjectedSmokeFeature) RegisterCollectors(FeatureContext, *prometheus.Registry) error {
	return nil
}

func (facadeInjectedSmokeFeature) SmokeSpec() SmokeSpec {
	return SmokeSpec{
		WantMetrics: []string{"facade_exporter_custom_metric 1"},
	}
}

func hasFacadeTestString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
