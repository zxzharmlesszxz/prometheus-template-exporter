package exporter

import (
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// Feature is the stable extension point for concrete exporters.
//
// A feature owns its domain flags and collector registration. Concrete exporter
// repositories add features in their own code and pass them to Main or RunCLI.
type Feature interface {
	RegisterFlags(app *kingpin.Application)
	RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error
}

type NamedFeature interface {
	FeatureName() string
}

type RuntimeConfigReporter interface {
	RuntimeConfig() []any
}

type DefaultListenAddressProvider interface {
	DefaultListenAddress() string
}

type FeatureContext struct {
	Logger       *slog.Logger
	ExporterName string
	Namespace    string
}

type CollectorFeature struct {
	Name                      string
	DefaultListenAddressValue string
	RegisterFlagsFunc         func(app *kingpin.Application)
	CollectorsFunc            func(ctx FeatureContext) ([]prometheus.Collector, error)
	RuntimeConfigFunc         func() []any
	RegisterCollectorsFunc    func(ctx FeatureContext, registry *prometheus.Registry) error
}

func (f CollectorFeature) FeatureName() string {
	return f.Name
}

func (f CollectorFeature) DefaultListenAddress() string {
	return f.DefaultListenAddressValue
}

func (f CollectorFeature) RegisterFlags(app *kingpin.Application) {
	if f.RegisterFlagsFunc != nil {
		f.RegisterFlagsFunc(app)
	}
}

func (f CollectorFeature) RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error {
	if f.RegisterCollectorsFunc != nil {
		return f.RegisterCollectorsFunc(ctx, registry)
	}
	if f.CollectorsFunc == nil {
		return nil
	}
	collectors, err := f.CollectorsFunc(ctx)
	if err != nil {
		return err
	}
	return RegisterCollectors(registry, collectors...)
}

func (f CollectorFeature) RuntimeConfig() []any {
	if f.RuntimeConfigFunc == nil {
		return nil
	}
	return f.RuntimeConfigFunc()
}

func RegisterCollectors(registry *prometheus.Registry, collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		if collector == nil {
			continue
		}
		if err := registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}
