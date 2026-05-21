package exporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileMTimeSeconds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	modTime := time.Unix(1_700_000_000, 123)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("set file times: %v", err)
	}

	if got := FileMTimeSeconds(path); got != float64(modTime.Unix()) {
		t.Fatalf("FileMTimeSeconds() = %v, want %v", got, modTime.Unix())
	}
}

func TestFileMTimeSecondsReturnsZeroForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")

	if got := FileMTimeSeconds(path); got != 0 {
		t.Fatalf("FileMTimeSeconds() = %v, want 0", got)
	}
}
