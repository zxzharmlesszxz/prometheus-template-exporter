package exporter

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
)

func TestParseCLIConfigBuildsOptionsAndRuntimeConfig(t *testing.T) {
	t.Parallel()

	feature := &cliTestFeature{}
	cfg := Config{
		Name:                 "cli_exporter",
		Namespace:            "cli_namespace",
		Description:          "CLI exporter",
		DefaultListenAddress: ":9777",
		DefaultMetricsPath:   "/metrics",
		Features:             []Feature{feature},
	}

	parsed, err := parseCLIConfig(cfg, []string{
		"--web.telemetry-path=/custom-metrics",
		"--web.enable-pprof",
		"--demo.value=custom",
	})
	if err != nil {
		t.Fatalf("parseCLIConfig() error = %v, want nil", err)
	}

	if parsed.options.Name != "cli_exporter" {
		t.Fatalf("Options.Name = %q, want %q", parsed.options.Name, "cli_exporter")
	}
	if parsed.options.Namespace != "cli_namespace" {
		t.Fatalf("Options.Namespace = %q, want %q", parsed.options.Namespace, "cli_namespace")
	}
	if parsed.options.Description != "CLI exporter" {
		t.Fatalf("Options.Description = %q, want %q", parsed.options.Description, "CLI exporter")
	}
	if parsed.options.MetricsPath != "/custom-metrics" {
		t.Fatalf("Options.MetricsPath = %q, want %q", parsed.options.MetricsPath, "/custom-metrics")
	}
	if !parsed.options.EnablePprof {
		t.Fatal("Options.EnablePprof = false, want true")
	}
	if parsed.options.ToolkitFlags == nil {
		t.Fatal("Options.ToolkitFlags = nil, want toolkit flags")
	}
	if parsed.promslogCfg == nil {
		t.Fatal("promslogCfg = nil, want config")
	}
	if len(parsed.options.Features) != 1 || parsed.options.Features[0] != feature {
		t.Fatalf("Options.Features = %v, want original feature", parsed.options.Features)
	}

	wantRuntimeConfig := []any{
		"metrics_path", "/custom-metrics",
		"pprof_enabled", true,
		"demo_value", "custom",
	}
	if !reflect.DeepEqual(parsed.runtimeConfig, wantRuntimeConfig) {
		t.Fatalf("runtimeConfig = %v, want %v", parsed.runtimeConfig, wantRuntimeConfig)
	}
}

func TestRunCLIServesConfiguredHandler(t *testing.T) {
	preserveVersionMetadata(t)

	feature := &cliIntegrationFeature{}
	cfg := Config{
		Name:                 "cli_exporter",
		Namespace:            "cli_exporter",
		Description:          "CLI exporter",
		DefaultListenAddress: ":9777",
		DefaultMetricsPath:   "/metrics",
		Features:             []Feature{feature},
	}

	called := false
	stubListenAndServe(t, func(srv *http.Server, flags *web.FlagConfig, logger *slog.Logger) error {
		called = true
		if logger == nil {
			t.Fatal("logger = nil, want logger")
		}
		if flags == nil {
			t.Fatal("ToolkitFlags = nil, want toolkit flags")
		}
		if flags.WebListenAddresses == nil || len(*flags.WebListenAddresses) != 1 || (*flags.WebListenAddresses)[0] != ":9888" {
			t.Fatalf("WebListenAddresses = %v, want [:9888]", flags.WebListenAddresses)
		}

		req := httptest.NewRequest(http.MethodGet, "/cli-metrics", nil)
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /cli-metrics status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Body.String(), "cli_integration_value 21") {
			t.Fatalf("GET /cli-metrics body missing feature metric: %s", rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/debug/pprof/heap?debug=1", nil)
		rec = httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /debug/pprof/heap status = %d, want %d", rec.Code, http.StatusOK)
		}
		return nil
	})

	err := RunCLI(cfg, []string{
		"--log.level=error",
		"--web.listen-address=:9888",
		"--web.telemetry-path=/cli-metrics",
		"--web.enable-pprof",
		"--demo.value=integration",
	})
	if err != nil {
		t.Fatalf("RunCLI() error = %v, want nil", err)
	}
	if !called {
		t.Fatal("listenAndServe was not called")
	}
}

func TestRunCLIReturnsInvalidTelemetryPathError(t *testing.T) {
	preserveVersionMetadata(t)

	err := RunCLI(Config{Name: "cli_exporter"}, []string{"--web.telemetry-path=metrics"})
	if err == nil {
		t.Fatal("RunCLI() error = nil, want invalid telemetry path error")
	}
	if !strings.Contains(err.Error(), `invalid --web.telemetry-path "metrics"`) {
		t.Fatalf("RunCLI() error = %q, want telemetry path context", err.Error())
	}
}

func TestRunCLIFromProjectReturnsParseError(t *testing.T) {
	preserveVersionMetadata(t)

	err := RunCLIFromProject([]string{"--not-a-real-flag"})
	if err == nil {
		t.Fatal("RunCLIFromProject() error = nil, want parse error")
	}
}

func preserveVersionMetadata(t *testing.T) {
	t.Helper()

	originalVersion := version.Version
	originalBranch := version.Branch
	originalRevision := version.Revision
	t.Cleanup(func() {
		version.Version = originalVersion
		version.Branch = originalBranch
		version.Revision = originalRevision
	})
}

type cliTestFeature struct {
	value *string
}

func (f *cliTestFeature) RegisterFlags(app *kingpin.Application) {
	f.value = app.Flag("demo.value", "Demo value").Default("default").String()
}

func (f *cliTestFeature) RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error {
	return nil
}

func (f *cliTestFeature) RuntimeConfig() []any {
	return []any{"demo_value", *f.value}
}

type cliIntegrationFeature struct {
	value *string
}

func (f *cliIntegrationFeature) RegisterFlags(app *kingpin.Application) {
	f.value = app.Flag("demo.value", "Demo value").Default("default").String()
}

func (f *cliIntegrationFeature) RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error {
	if f.value == nil {
		return fmt.Errorf("demo.value flag was not registered")
	}
	if *f.value != "integration" {
		return fmt.Errorf("demo.value = %q, want integration", *f.value)
	}
	return RegisterCollectors(registry, newConstCollector("cli_integration_value", "CLI integration value", 21))
}
