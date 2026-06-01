package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/exporter-toolkit/web"
)

type Options struct {
	Name         string
	Namespace    string
	Description  string
	MetricsPath  string
	ToolkitFlags *web.FlagConfig
	EnablePprof  bool
	Features     []Feature
}

var listenAndServe = web.ListenAndServe

func (o Options) normalized() Options {
	if o.Name == "" {
		o.Name = defaultExporterName
	}
	if o.Namespace == "" {
		o.Namespace = o.Name
	}
	if o.Description == "" {
		o.Description = defaultDescription
	}
	if o.MetricsPath == "" {
		o.MetricsPath = defaultTelemetryPath
	}
	return o
}

func Run(opts Options, logger *slog.Logger) error {
	opts = opts.normalized()
	if err := validateMetricsPath(opts.MetricsPath); err != nil {
		return fmt.Errorf("invalid metrics path %q: %w", opts.MetricsPath, err)
	}
	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.With("component", "server")
	registry, err := NewRegistry(opts.Namespace, logger.With("component", "collector"), opts.Features...)
	if err != nil {
		return fmt.Errorf("create registry: %w", err)
	}

	handler := NewHandler(HandlerOptions{
		Name:        opts.Name,
		Description: opts.Description,
		MetricsPath: opts.MetricsPath,
		Registry:    registry,
		EnablePprof: opts.EnablePprof,
	})

	srv := &http.Server{Handler: handler}
	return listenAndServe(srv, opts.ToolkitFlags, logger)
}

func MustRun(opts Options, logger *slog.Logger) {
	if err := Run(opts, logger); err != nil {
		panic(err)
	}
}

func NewServer(opts Options, registry *prometheus.Registry) *http.Server {
	opts = opts.normalized()
	return &http.Server{
		Handler: NewHandler(HandlerOptions{
			Name:        opts.Name,
			Description: opts.Description,
			MetricsPath: opts.MetricsPath,
			Registry:    registry,
			EnablePprof: opts.EnablePprof,
		}),
	}
}

func NewServerChecked(opts Options, registry *prometheus.Registry) (*http.Server, error) {
	opts = opts.normalized()
	handler, err := NewHandlerChecked(HandlerOptions{
		Name:        opts.Name,
		Description: opts.Description,
		MetricsPath: opts.MetricsPath,
		Registry:    registry,
		EnablePprof: opts.EnablePprof,
	})
	if err != nil {
		return nil, err
	}
	return &http.Server{Handler: handler}, nil
}
