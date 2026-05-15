package exporter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestHandlerServesMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	if err := RegisterCollectors(registry, newConstCollector("template_example_metric", "Example metric", 7)); err != nil {
		t.Fatalf("RegisterCollectors() error = %v", err)
	}

	handler := NewHandler(HandlerOptions{
		Name:        "template_exporter",
		Description: "Template exporter",
		MetricsPath: "/metrics",
		Registry:    registry,
		EnablePprof: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "template_example_metric 7") {
		t.Fatalf("GET /metrics body missing custom metric: %s", rec.Body.String())
	}
}

func TestHandlerServesHealthz(t *testing.T) {
	t.Parallel()

	handler := NewHandler(HandlerOptions{Registry: prometheus.NewRegistry()})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("GET /healthz body = %q, want %q", rec.Body.String(), "ok\n")
	}
}

func TestHandlerServesLandingPage(t *testing.T) {
	t.Parallel()

	handler := NewHandler(HandlerOptions{
		Name:        "template_exporter",
		Description: "Reusable Prometheus exporter template",
		MetricsPath: "/metrics",
		Registry:    prometheus.NewRegistry(),
		EnablePprof: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"Reusable Prometheus exporter template",
		"/metrics",
		"/healthz",
		"/debug/pprof/heap",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET / body missing %q: %s", want, body)
		}
	}
}

func TestHandlerDisablesPprofWhenConfigured(t *testing.T) {
	t.Parallel()

	handler := NewHandler(HandlerOptions{
		MetricsPath: "/metrics",
		Registry:    prometheus.NewRegistry(),
		EnablePprof: false,
	})

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/heap?debug=1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /debug/pprof/heap status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
