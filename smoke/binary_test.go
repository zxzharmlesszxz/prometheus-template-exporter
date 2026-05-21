package smoke

import (
	"testing"

	"github.com/zxzharmlesszxz/prometheus-exporter-framework/exporter/exportertest/smoketest"
)

func TestBinarySmoke(t *testing.T) {
	smoketest.RunBinary(t, smoketest.Config{
		ProjectName:       "prometheus-exporter-framework",
		BuildInfoMetric:   "exporter_framework_build_info",
		RenamedExecutable: "renamed-exporter-framework",
	})
}
