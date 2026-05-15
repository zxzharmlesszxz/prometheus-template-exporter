# Architecture

## Overview

`prometheus-template-exporter` is a reusable exporter shell.

It intentionally does not know about any business domain such as Debian packages, Puppet agent files, or Puppetfile parsing.
Domain logic is supplied by external `exporter.Feature` implementations.

## Package Layout

- `cmd`
  Minimal binary that runs the shell without domain features.
- `exporter`
  Public framework package used by concrete exporters.

## Data Flow

1. A concrete exporter calls `exporter.MainFromProject(feature)` or `exporter.Main(exporter.Config{...})` when it needs explicit metadata overrides.
2. The template registers common CLI flags and asks each feature to register its own flags.
3. After parsing, the template creates the logger, runtime options, and Prometheus registry.
4. The registry always receives:
   - `*_build_info`
   - Go runtime collectors
   - process collectors
5. Each feature registers its own collectors.
6. The HTTP server exposes the registry on the configured telemetry path.

## Extension Contract

Features implement:

```go
type Feature interface {
	RegisterFlags(app *kingpin.Application)
	RegisterCollectors(ctx FeatureContext, registry *prometheus.Registry) error
}
```

Optional interfaces:

- `NamedFeature`
  Adds a feature name to structured logs and registration errors.
- `RuntimeConfigReporter`
  Adds feature-specific fields to the startup runtime config log.
- `DefaultListenAddressProvider`
  Provides the feature's default `--web.listen-address` value.

The helper `CollectorFeature` can be used when a feature only needs callbacks instead of a dedicated type.

## Common HTTP Semantics

- `/metrics` exposes the configured registry.
- `/healthz` returns `200 OK` with `ok\n` while the process is serving requests.
- `/debug/pprof/*` is disabled unless `--web.enable-pprof` is set.
- The landing page links to metrics and health endpoints.

`/healthz` reflects process health only.
Domain-source health belongs in feature collectors.

## Metric Ownership

The template owns only common exporter metrics:

- `*_build_info`
- Go runtime metrics
- process metrics

Every business metric and domain diagnostic metric is owned by the concrete feature.
For example:

- Debian package metrics belong to the package feature.
- Puppet agent source-health metrics belong to the Puppet feature.
- Puppetfile module metrics belong to the Puppetfile feature.

## Failure Semantics

The template fails startup when a feature cannot register its collectors.
Runtime scrape failures are domain-specific and should be represented by feature collectors.

For example, a feature can:

- emit `*_up = 0`
- serve cached metrics after a failed refresh
- omit business metrics when its source is unreadable
- increment read or parse error counters

The template does not impose one policy because the existing exporters use different failure semantics.
