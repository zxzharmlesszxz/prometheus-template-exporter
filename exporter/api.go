package exporter

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/app"
	"github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/feature"
	"github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/files"
	"github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/metric"
	snapshotpkg "github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/internal/snapshot"
)

type Config = app.Config

func ConfigFromProject(features ...Feature) Config {
	return app.ConfigFromProject(features...)
}

func ConfigForProject(projectName string, features ...Feature) Config {
	return app.ConfigForProject(projectName, features...)
}

func ExporterNameFromProject(projectName string) string {
	return app.ExporterNameFromProject(projectName)
}

func DescriptionFromProject(projectName string) string {
	return app.DescriptionFromProject(projectName)
}

func Main(cfg Config) {
	app.Main(cfg)
}

func MainFromProject(features ...Feature) {
	app.MainFromProject(features...)
}

func MainForProject(projectName, description string, features ...Feature) {
	app.MainForProject(projectName, description, features...)
}

func RunCLIFromProject(args []string, features ...Feature) error {
	return app.RunCLIFromProject(args, features...)
}

func RunCLI(cfg Config, args []string) error {
	return app.RunCLI(cfg, args)
}

type Feature = feature.Feature
type NamedFeature = feature.NamedFeature
type RuntimeConfigReporter = feature.RuntimeConfigReporter
type SmokeSpecProvider = feature.SmokeSpecProvider
type SmokeSpec = feature.SmokeSpec
type DefaultListenAddressProvider = feature.DefaultListenAddressProvider
type StartableCollector = feature.StartableCollector
type FeatureContext = feature.FeatureContext
type CollectorFeature = feature.CollectorFeature

func RegisterCollectors(registry *prometheus.Registry, collectors ...prometheus.Collector) error {
	return feature.RegisterCollectors(registry, collectors...)
}

func RegisterAndStartCollectors(ctx context.Context, registry *prometheus.Registry, collectors ...StartableCollector) error {
	return feature.RegisterAndStartCollectors(ctx, registry, collectors...)
}

func FileMTimeSeconds(path string) float64 {
	return files.FileMTimeSeconds(path)
}

type Uint64Counter = files.Uint64Counter
type FileReadFunc = files.FileReadFunc
type FileScrapeResult = files.FileScrapeResult
type FileScraper = files.FileScraper
type FileScrapeMetrics = files.FileScrapeMetrics

type HandlerOptions = app.HandlerOptions

func NewHandler(opts HandlerOptions) http.Handler {
	return app.NewHandler(opts)
}

func NewHandlerChecked(opts HandlerOptions) (http.Handler, error) {
	return app.NewHandlerChecked(opts)
}

type ProjectMetadata = app.ProjectMetadata
type ExporterInfo = app.ExporterInfo
type MetricInfo = app.MetricInfo
type SmokeInfo = app.SmokeInfo

var (
	injectedExporterName        string
	injectedExporterDescription string
	injectedFeatureName         string
	injectedMetricNamespace     string
	injectedListenAddress       string
)

func InjectedExporterName() string {
	return requireInjectedDefault("injectedExporterName", injectedExporterName)
}

func InjectedExporterDescription() string {
	return requireInjectedDefault("injectedExporterDescription", injectedExporterDescription)
}

func InjectedFeatureName() string {
	return requireInjectedDefault("injectedFeatureName", injectedFeatureName)
}

func InjectedMetricNamespace() string {
	return requireInjectedDefault("injectedMetricNamespace", injectedMetricNamespace)
}

func InjectedDefaultListenAddress() string {
	listenAddress := requireInjectedDefault("injectedListenAddress", injectedListenAddress)
	requireListenAddress(listenAddress)
	return listenAddress
}

func InjectedProjectMetadata() ProjectMetadata {
	return ProjectMetadata{
		ExporterName:         InjectedExporterName(),
		ExporterDescription:  InjectedExporterDescription(),
		FeatureName:          InjectedFeatureName(),
		MetricNamespace:      InjectedMetricNamespace(),
		DefaultListenAddress: InjectedDefaultListenAddress(),
	}
}

func ConfigFromInjectedProject(features ...Feature) Config {
	metadata := InjectedProjectMetadata()
	return Config{
		Name:                 metadata.ExporterName,
		Namespace:            metadata.MetricNamespace,
		Description:          metadata.ExporterDescription,
		DefaultListenAddress: metadata.DefaultListenAddress,
		Features:             features,
	}
}

func MainFromInjectedProject(features ...Feature) {
	cfg := ConfigFromInjectedProject(features...)
	cfg.Name = executableName(os.Args, cfg.Name)
	Main(cfg)
}

func ExporterInfoFromInjectedProject(features ...Feature) ExporterInfo {
	return ExporterInfoFromProjectMetadata(InjectedProjectMetadata(), features...)
}

func ExporterInfoFromProjectMetadata(metadata ProjectMetadata, features ...Feature) ExporterInfo {
	return app.ExporterInfoFromProjectMetadata(metadata, features...)
}

func StandardMetricInfo(namespace string) MetricInfo {
	return app.StandardMetricInfo(namespace)
}

func NewRegistry(namespace string, logger *slog.Logger, features ...Feature) (*prometheus.Registry, error) {
	return app.NewRegistry(namespace, logger, features...)
}

type Options = app.Options

func Run(opts Options, logger *slog.Logger) error {
	return app.Run(opts, logger)
}

func MustRun(opts Options, logger *slog.Logger) {
	app.MustRun(opts, logger)
}

func NewServer(opts Options, registry *prometheus.Registry) *http.Server {
	return app.NewServer(opts, registry)
}

func NewServerChecked(opts Options, registry *prometheus.Registry) (*http.Server, error) {
	return app.NewServerChecked(opts, registry)
}

const DefaultSnapshotRefreshInterval = snapshotpkg.DefaultSnapshotRefreshInterval

type Snapshotter[T any] = snapshotpkg.Snapshotter[T]
type SnapshotStatus = snapshotpkg.SnapshotStatus
type SnapshotCollectorOptions[T any] = snapshotpkg.SnapshotCollectorOptions[T]
type SnapshotCollector[T any] = snapshotpkg.SnapshotCollector[T]

func NewSnapshotCollector[T any](options SnapshotCollectorOptions[T]) *SnapshotCollector[T] {
	return snapshotpkg.NewSnapshotCollector(options)
}

func HydrateVersionMetadata() {
	app.HydrateVersionMetadata()
}

func ResolveVersionMetadata(currentVersion string, currentBranch string, currentRevision string, computedRevision string, buildMainVersion string, buildBranch string, buildRevision string) (string, string, string) {
	return app.ResolveVersionMetadata(currentVersion, currentBranch, currentRevision, computedRevision, buildMainVersion, buildBranch, buildRevision)
}

func BoolFloat(value bool) float64 {
	return metric.BoolFloat(value)
}

func UnixTimestamp(value time.Time) float64 {
	return metric.UnixTimestamp(value)
}

func NormalizeDuration(value time.Duration, fallback time.Duration) time.Duration {
	return metric.NormalizeDuration(value, fallback)
}

func executableName(args []string, fallback string) string {
	if len(args) == 0 {
		return fallback
	}

	name := filepath.Base(args[0])
	if name == "." || strings.TrimSpace(name) == "" {
		return fallback
	}
	return name
}

func requireInjectedDefault(name string, value string) string {
	if strings.TrimSpace(value) == "" {
		panic("missing Makefile-injected exporter metadata: " + name)
	}
	return value
}

func requireListenAddress(value string) {
	if !strings.HasPrefix(value, ":") {
		panic("invalid Makefile-injected exporter metadata: default listen address must start with :")
	}
}
