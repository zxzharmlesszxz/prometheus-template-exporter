package exportertest

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}

func Register(tb TB, registry *prometheus.Registry, collector prometheus.Collector) {
	tb.Helper()

	if err := registry.Register(collector); err != nil {
		tb.Fatalf("register collector: %v", err)
	}
}

func Gather(tb TB, gatherer prometheus.Gatherer) []*dto.MetricFamily {
	tb.Helper()

	families, err := gatherer.Gather()
	if err != nil {
		tb.Fatalf("gather metrics: %v", err)
	}
	return families
}

func RegisterAndGather(tb TB, collector prometheus.Collector) []*dto.MetricFamily {
	tb.Helper()

	registry := prometheus.NewRegistry()
	Register(tb, registry, collector)
	return Gather(tb, registry)
}

func MetricFamily(tb TB, families []*dto.MetricFamily, name string) *dto.MetricFamily {
	tb.Helper()

	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	tb.Fatalf("metric family %q not found", name)
	return nil
}

func MetricValue(families []*dto.MetricFamily, name string, labels map[string]string) (float64, bool) {
	family := metricFamily(families, name)
	if family == nil {
		return 0, false
	}
	for _, metric := range family.GetMetric() {
		if !LabelsMatch(metric, labels) {
			continue
		}
		switch {
		case metric.Gauge != nil:
			return metric.GetGauge().GetValue(), true
		case metric.Counter != nil:
			return metric.GetCounter().GetValue(), true
		case metric.Untyped != nil:
			return metric.GetUntyped().GetValue(), true
		}
	}
	return 0, false
}

func AssertMetricValue(tb TB, families []*dto.MetricFamily, name string, labels map[string]string, want float64) {
	tb.Helper()

	got, ok := MetricValue(families, name, labels)
	if !ok {
		tb.Fatalf("metric %s%v not found", name, labels)
	}
	if got != want {
		tb.Fatalf("%s%v = %v, want %v", name, labels, got, want)
	}
}

func RuntimeConfigValue(tb TB, config []any, key string) any {
	tb.Helper()

	for i := 0; i+1 < len(config); i += 2 {
		if config[i] == key {
			return config[i+1]
		}
	}
	tb.Fatalf("missing runtime config key %q in %#v", key, config)
	return nil
}

func WaitForMetricValue(tb TB, gatherer prometheus.Gatherer, name string, labels map[string]string, want float64) {
	tb.Helper()

	deadline := time.Now().Add(time.Second)
	for {
		families := Gather(tb, gatherer)
		got, ok := MetricValue(families, name, labels)
		if ok && got == want {
			return
		}
		if time.Now().After(deadline) {
			tb.Fatalf("%s%v did not become %v", name, labels, want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func Histogram(tb TB, families []*dto.MetricFamily, name string, labels map[string]string) *dto.Histogram {
	tb.Helper()

	family := metricFamily(families, name)
	if family == nil {
		tb.Fatalf("histogram family %q not found", name)
	}
	for _, metric := range family.GetMetric() {
		if LabelsMatch(metric, labels) && metric.Histogram != nil {
			return metric.GetHistogram()
		}
	}
	tb.Fatalf("histogram %s%v not found", name, labels)
	return nil
}

func LabelsMatch(metric *dto.Metric, want map[string]string) bool {
	for name, value := range want {
		found := false
		for _, label := range metric.GetLabel() {
			if label.GetName() == name && label.GetValue() == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func metricFamily(families []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}
