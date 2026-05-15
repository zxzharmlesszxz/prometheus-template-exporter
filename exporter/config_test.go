package exporter

import "testing"

func TestConfigNormalizedSetsDefaults(t *testing.T) {
	t.Parallel()

	cfg := Config{}.normalized()

	if cfg.Name != defaultExporterName {
		t.Fatalf("Name = %q, want %q", cfg.Name, defaultExporterName)
	}
	if cfg.Namespace != defaultExporterName {
		t.Fatalf("Namespace = %q, want %q", cfg.Namespace, defaultExporterName)
	}
	if cfg.Description != defaultDescription {
		t.Fatalf("Description = %q, want %q", cfg.Description, defaultDescription)
	}
	if cfg.DefaultListenAddress != defaultListenAddress {
		t.Fatalf("DefaultListenAddress = %q, want %q", cfg.DefaultListenAddress, defaultListenAddress)
	}
	if cfg.DefaultMetricsPath != defaultTelemetryPath {
		t.Fatalf("DefaultMetricsPath = %q, want %q", cfg.DefaultMetricsPath, defaultTelemetryPath)
	}
}

func TestConfigNormalizedUsesNameAsNamespace(t *testing.T) {
	t.Parallel()

	cfg := Config{Name: "custom_exporter"}.normalized()
	if cfg.Namespace != "custom_exporter" {
		t.Fatalf("Namespace = %q, want %q", cfg.Namespace, "custom_exporter")
	}
}

func TestConfigForProject(t *testing.T) {
	t.Parallel()

	feature := CollectorFeature{Name: "demo", DefaultListenAddressValue: ":9888"}
	cfg := ConfigForProject("git.example.net/platform/prometheus-pkg-exporter", feature)

	if cfg.Name != "pkg_exporter" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "pkg_exporter")
	}
	if cfg.Namespace != "pkg_exporter" {
		t.Fatalf("Namespace = %q, want %q", cfg.Namespace, "pkg_exporter")
	}
	if cfg.Description != "Prometheus Pkg Exporter" {
		t.Fatalf("Description = %q, want %q", cfg.Description, "Prometheus Pkg Exporter")
	}
	if cfg.DefaultListenAddress != ":9888" {
		t.Fatalf("DefaultListenAddress = %q, want %q", cfg.DefaultListenAddress, ":9888")
	}
	if len(cfg.Features) != 1 {
		t.Fatalf("Features len = %d, want 1", len(cfg.Features))
	}
}

func TestConfigForProjectFallsBackToDefaultListenAddress(t *testing.T) {
	t.Parallel()

	cfg := ConfigForProject("prometheus-pkg-exporter")
	if cfg.DefaultListenAddress != defaultListenAddress {
		t.Fatalf("DefaultListenAddress = %q, want %q", cfg.DefaultListenAddress, defaultListenAddress)
	}
}

func TestExporterNameFromProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		project string
		want    string
	}{
		{project: "prometheus-pkg-exporter", want: "pkg_exporter"},
		{project: "example.com/team/prometheus-puppetfile-exporter", want: "puppetfile_exporter"},
		{project: "custom-exporter", want: "custom_exporter"},
		{project: "123-custom", want: "_123_custom_exporter"},
		{project: "", want: "template_exporter"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.project, func(t *testing.T) {
			t.Parallel()

			if got := ExporterNameFromProject(tc.project); got != tc.want {
				t.Fatalf("ExporterNameFromProject(%q) = %q, want %q", tc.project, got, tc.want)
			}
		})
	}
}
