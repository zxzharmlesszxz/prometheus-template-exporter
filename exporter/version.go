package exporter

import (
	"runtime/debug"
	"strings"

	"github.com/prometheus/common/version"
)

func HydrateVersionMetadata() {
	mainVersion, branch, revision := readBuildInfoMetadata()
	resolvedVersion, resolvedBranch, resolvedRevision := ResolveVersionMetadata(
		version.Version,
		version.Branch,
		version.Revision,
		version.GetRevision(),
		mainVersion,
		branch,
		revision,
	)
	version.Version = resolvedVersion
	version.Branch = resolvedBranch
	version.Revision = resolvedRevision
}

func readBuildInfoMetadata() (string, string, string) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", ""
	}

	var branch string
	var revision string
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.branch":
			branch = setting.Value
		case "vcs.revision":
			revision = setting.Value
		}
	}

	return buildInfo.Main.Version, branch, revision
}

func ResolveVersionMetadata(currentVersion string, currentBranch string, currentRevision string, computedRevision string, buildMainVersion string, buildBranch string, buildRevision string) (string, string, string) {
	resolvedVersion := firstNonEmpty(currentVersion, normalizeBuildVersion(buildMainVersion), "dev")
	resolvedBranch := firstNonEmpty(currentBranch, strings.TrimSpace(buildBranch), "dev")
	resolvedRevision := firstNonEmpty(currentRevision, normalizeComputedRevision(buildRevision), normalizeComputedRevision(computedRevision), "dev")
	return resolvedVersion, resolvedBranch, resolvedRevision
}

func normalizeBuildVersion(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "(devel)" {
		return ""
	}
	return trimmed
}

func normalizeComputedRevision(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "unknown" {
		return ""
	}
	return trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
