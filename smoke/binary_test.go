package smoke

import (
	"testing"

	"github.com/zxzharmlesszxz/prometheus-template-exporter/exporter/exportertest/smoketest"
)

func TestBinarySmoke(t *testing.T) {
	smoketest.RunBinary(t, smoketest.Config{
		ProjectName:       "prometheus-template-exporter",
		BuildInfoMetric:   "template_exporter_build_info",
		RenamedExecutable: "renamed-template-exporter",
	})
}
