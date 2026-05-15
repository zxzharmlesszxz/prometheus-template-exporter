package exporter

import "testing"

func TestResolveVersionMetadata(t *testing.T) {
	t.Parallel()

	t.Run("uses current values when provided", func(t *testing.T) {
		t.Parallel()

		v, b, r := ResolveVersionMetadata("1.2.3", "main", "abc123", "ignored", "v9.9.9", "other", "def456")
		if v != "1.2.3" || b != "main" || r != "abc123" {
			t.Fatalf("ResolveVersionMetadata() = (%q, %q, %q), want (%q, %q, %q)", v, b, r, "1.2.3", "main", "abc123")
		}
	})

	t.Run("uses build info when current values are empty", func(t *testing.T) {
		t.Parallel()

		v, b, r := ResolveVersionMetadata("", "", "", "ignored", "v0.4.0", "release", "deadbeef")
		if v != "v0.4.0" || b != "release" || r != "deadbeef" {
			t.Fatalf("ResolveVersionMetadata() = (%q, %q, %q), want (%q, %q, %q)", v, b, r, "v0.4.0", "release", "deadbeef")
		}
	})

	t.Run("uses computed revision when build revision is unknown", func(t *testing.T) {
		t.Parallel()

		v, b, r := ResolveVersionMetadata("", "", "", "f00baa", "(devel)", "", "unknown")
		if v != "dev" || b != "dev" || r != "f00baa" {
			t.Fatalf("ResolveVersionMetadata() = (%q, %q, %q), want (%q, %q, %q)", v, b, r, "dev", "dev", "f00baa")
		}
	})

	t.Run("falls back to dev when no metadata exists", func(t *testing.T) {
		t.Parallel()

		v, b, r := ResolveVersionMetadata("", "", "", "unknown", "", "", "")
		if v != "dev" || b != "dev" || r != "dev" {
			t.Fatalf("ResolveVersionMetadata() = (%q, %q, %q), want (%q, %q, %q)", v, b, r, "dev", "dev", "dev")
		}
	})
}
