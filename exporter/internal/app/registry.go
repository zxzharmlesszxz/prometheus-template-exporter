package app

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
)

func NewRegistry(namespace string, logger *slog.Logger, features ...Feature) (*prometheus.Registry, error) {
	if namespace == "" {
		namespace = defaultExporterName
	}
	if logger == nil {
		logger = slog.Default()
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(versioncollector.NewCollector(namespace))
	registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	for _, feature := range features {
		if feature == nil {
			continue
		}

		featureLogger := logger
		if name := featureName(feature); name != "" {
			featureLogger = featureLogger.With("feature", name)
		}

		ctx := FeatureContext{
			Logger:       featureLogger,
			ExporterName: namespace,
			Namespace:    namespace,
		}
		if err := feature.RegisterCollectors(ctx, registry); err != nil {
			return nil, fmt.Errorf("register feature %q: %w", featureName(feature), err)
		}
	}

	return registry, nil
}

func featureName(feature Feature) string {
	named, ok := feature.(NamedFeature)
	if !ok {
		return ""
	}
	return named.FeatureName()
}
