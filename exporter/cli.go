package exporter

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	promflag "github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

func Main(cfg Config) {
	if err := RunCLI(cfg, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func MainFromProject(features ...Feature) {
	Main(ConfigFromProject(features...))
}

func RunCLIFromProject(args []string, features ...Feature) error {
	return RunCLI(ConfigFromProject(features...), args)
}

func RunCLI(cfg Config, args []string) error {
	cfg = cfg.normalized()
	HydrateVersionMetadata()

	app := kingpin.New(cfg.Name, cfg.Description)
	promslogCfg := &promslog.Config{}
	promflag.AddFlags(app, promslogCfg)

	toolkitFlags := webflag.AddFlags(app, cfg.DefaultListenAddress)
	metricsPath := app.Flag(
		"web.telemetry-path", "Path under which to expose metrics",
	).Default(cfg.DefaultMetricsPath).String()
	enablePprof := app.Flag(
		"web.enable-pprof", "Expose pprof endpoints and links on the landing page",
	).Default("false").Bool()

	for _, feature := range cfg.Features {
		if feature != nil {
			feature.RegisterFlags(app)
		}
	}

	app.Version(version.Print(cfg.Namespace))
	app.HelpFlag.Short('h')
	if _, err := app.Parse(args); err != nil {
		return err
	}

	logger := promslog.New(promslogCfg)
	logger.Info("Starting "+cfg.Name, "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	opts := Options{
		Name:         cfg.Name,
		Namespace:    cfg.Namespace,
		Description:  cfg.Description,
		MetricsPath:  *metricsPath,
		ToolkitFlags: toolkitFlags,
		EnablePprof:  *enablePprof,
		Features:     cfg.Features,
	}

	runtimeConfig := []any{
		"metrics_path", opts.MetricsPath,
		"pprof_enabled", opts.EnablePprof,
	}
	for _, feature := range cfg.Features {
		reporter, ok := feature.(RuntimeConfigReporter)
		if !ok {
			continue
		}
		runtimeConfig = append(runtimeConfig, reporter.RuntimeConfig()...)
	}
	logger.Info("Runtime config", runtimeConfig...)

	return Run(opts, logger)
}
