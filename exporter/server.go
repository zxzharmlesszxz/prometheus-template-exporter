package exporter

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
	return web.ListenAndServe(srv, opts.ToolkitFlags, logger)
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
