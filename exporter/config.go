package exporter

import (
	"path"
	"runtime/debug"
	"strings"
)

const (
	defaultExporterName   = "template_exporter"
	defaultDescription    = "Prometheus exporter template"
	defaultListenAddress  = ":9900"
	defaultTelemetryPath  = "/metrics"
	defaultLandingName    = "template_exporter"
	defaultProfilingValue = "false"
)

type Config struct {
	Name                 string
	Namespace            string
	Description          string
	DefaultListenAddress string
	DefaultMetricsPath   string
	Features             []Feature
}

func ConfigFromProject(features ...Feature) Config {
	return ConfigForProject(moduleProjectName(), features...)
}

func ConfigForProject(projectName string, features ...Feature) Config {
	name := ExporterNameFromProject(projectName)
	return Config{
		Name:                 name,
		Namespace:            name,
		Description:          DescriptionFromProject(projectName),
		DefaultListenAddress: defaultListenAddressFromFeatures(features),
		Features:             features,
	}
}

func ExporterNameFromProject(projectName string) string {
	base := projectDomainName(projectName)
	name := sanitizeMetricNamespace(base)
	if name == "" {
		return defaultExporterName
	}
	if !strings.HasSuffix(name, "_exporter") {
		name += "_exporter"
	}
	return name
}

func DescriptionFromProject(projectName string) string {
	base := projectBase(projectName)
	if base == "" {
		return defaultDescription
	}

	parts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	if len(parts) == 0 {
		return defaultDescription
	}

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func (c Config) normalized() Config {
	if c.Name == "" {
		c.Name = defaultExporterName
	}
	if c.Namespace == "" {
		c.Namespace = c.Name
	}
	if c.Description == "" {
		c.Description = defaultDescription
	}
	if c.DefaultListenAddress == "" {
		c.DefaultListenAddress = defaultListenAddress
	}
	if c.DefaultMetricsPath == "" {
		c.DefaultMetricsPath = defaultTelemetryPath
	}
	return c
}

func moduleProjectName() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return buildInfo.Main.Path
}

func projectDomainName(projectName string) string {
	base := projectBase(projectName)
	base = strings.TrimPrefix(base, "prometheus-")
	base = strings.TrimSuffix(base, "-exporter")
	base = strings.TrimSuffix(base, "_exporter")
	return base
}

func projectBase(projectName string) string {
	base := strings.TrimSpace(projectName)
	if base == "" {
		return ""
	}
	return path.Base(base)
}

func sanitizeMetricNamespace(value string) string {
	var builder strings.Builder
	lastUnderscore := false

	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastUnderscore = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return ""
	}
	if result[0] >= '0' && result[0] <= '9' {
		return "_" + result
	}
	return result
}

func defaultListenAddressFromFeatures(features []Feature) string {
	for _, feature := range features {
		provider, ok := feature.(DefaultListenAddressProvider)
		if !ok {
			continue
		}
		listenAddress := strings.TrimSpace(provider.DefaultListenAddress())
		if listenAddress != "" {
			return listenAddress
		}
	}
	return defaultListenAddress
}
