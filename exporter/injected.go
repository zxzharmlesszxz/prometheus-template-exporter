package exporter

import (
	"os"
	"strings"
)

var (
	injectedExporterName        string
	injectedExporterDescription string
	injectedFeatureName         string
	injectedMetricNamespace     string
	injectedListenAddress       string
)

type ProjectMetadata struct {
	ExporterName         string
	ExporterDescription  string
	FeatureName          string
	MetricNamespace      string
	DefaultListenAddress string
}

type ExporterInfo struct {
	Name                 string
	Description          string
	FeatureName          string
	MetricNamespace      string
	DefaultListenAddress string
	Metrics              MetricInfo
	Smoke                SmokeInfo
}

type MetricInfo struct {
	BuildInfo                                string
	LastCollectionSuccess                    string
	LastCollectionTimestampSeconds           string
	LastSuccessfulCollectionTimestampSeconds string
}

type SmokeInfo struct {
	ForbiddenUsageNames []string
	RenamedExecutable   string
	ServerArgs          []string
	WantMetrics         []string
	RejectMetrics       []string
}

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
	metadata.requireValid()
	metrics := StandardMetricInfo(metadata.MetricNamespace)
	smoke := SmokeInfo{
		ForbiddenUsageNames: []string{metadata.MetricNamespace},
		RenamedExecutable:   "renamed-" + metadata.FeatureName + "-exporter",
		ServerArgs:          []string{"--" + metadata.FeatureName + ".refresh-interval=100ms"},
		WantMetrics:         []string{metrics.LastCollectionSuccess + " 1"},
		RejectMetrics:       []string{metrics.LastCollectionSuccess + " 0"},
	}
	for _, feature := range features {
		provider, ok := feature.(SmokeSpecProvider)
		if !ok {
			continue
		}
		smoke = appendSmokeSpec(smoke, provider.SmokeSpec())
	}
	return ExporterInfo{
		Name:                 metadata.ExporterName,
		Description:          metadata.ExporterDescription,
		FeatureName:          metadata.FeatureName,
		MetricNamespace:      metadata.MetricNamespace,
		DefaultListenAddress: metadata.DefaultListenAddress,
		Metrics:              metrics,
		Smoke:                smoke,
	}
}

func StandardMetricInfo(namespace string) MetricInfo {
	return MetricInfo{
		BuildInfo:                                namespace + "_build_info",
		LastCollectionSuccess:                    namespace + "_last_collection_success",
		LastCollectionTimestampSeconds:           namespace + "_last_collection_timestamp_seconds",
		LastSuccessfulCollectionTimestampSeconds: namespace + "_last_successful_collection_timestamp_seconds",
	}
}

func appendSmokeSpec(info SmokeInfo, spec SmokeSpec) SmokeInfo {
	info.ServerArgs = append(info.ServerArgs, spec.ServerArgs...)
	info.WantMetrics = append(info.WantMetrics, spec.WantMetrics...)
	info.RejectMetrics = append(info.RejectMetrics, spec.RejectMetrics...)
	return info
}

func (m ProjectMetadata) requireValid() {
	requireInjectedDefault("ProjectMetadata.ExporterName", m.ExporterName)
	requireInjectedDefault("ProjectMetadata.ExporterDescription", m.ExporterDescription)
	requireInjectedDefault("ProjectMetadata.FeatureName", m.FeatureName)
	requireInjectedDefault("ProjectMetadata.MetricNamespace", m.MetricNamespace)
	requireInjectedDefault("ProjectMetadata.DefaultListenAddress", m.DefaultListenAddress)
	requireListenAddress(m.DefaultListenAddress)
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
