package smoke

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	smokeVersion   = "v9.8.7"
	smokeBranch    = "smoke-branch"
	smokeRevision  = "abc123def"
	smokeBuildUser = "smoke-test"
	smokeBuildDate = "2026-05-17T00:00:00Z"
)

func TestBinarySmoke(t *testing.T) {
	if os.Getenv("RUN_BINARY_SMOKE") != "1" {
		t.Skip("set RUN_BINARY_SMOKE=1 to run binary smoke test")
	}

	root := repoRoot(t)
	binary := buildBinary(t, root)

	t.Run("prints injected version metadata", func(t *testing.T) {
		output := runBinary(t, binary, "--version")
		for _, want := range []string{smokeVersion, smokeBranch, smokeRevision} {
			if !strings.Contains(output, want) {
				t.Fatalf("--version output missing %q: %s", want, output)
			}
		}
	})

	t.Run("rejects invalid telemetry path", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "--web.telemetry-path=metrics")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("invalid telemetry path exited successfully, output: %s", output)
		}
		if !strings.Contains(string(output), `invalid --web.telemetry-path "metrics"`) {
			t.Fatalf("invalid telemetry path output missing validation error: %s", output)
		}
	})

	t.Run("serves health and metrics", func(t *testing.T) {
		addr := freeAddress(t)
		baseURL := "http://" + addr
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := exec.CommandContext(
			ctx,
			binary,
			"--log.level=error",
			"--web.listen-address="+addr,
			"--web.telemetry-path=/metrics",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			t.Fatalf("start binary: %v", err)
		}
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- cmd.Wait()
			close(waitCh)
		}()
		t.Cleanup(func() {
			cancel()
			select {
			case <-waitCh:
			case <-time.After(5 * time.Second):
				_ = cmd.Process.Kill()
				<-waitCh
			}
		})

		client := &http.Client{Timeout: 500 * time.Millisecond}
		waitForHealth(t, client, baseURL, waitCh, &stdout, &stderr)

		health := httpGet(t, client, baseURL+"/healthz", http.StatusOK)
		if health != "ok\n" {
			t.Fatalf("GET /healthz body = %q, want %q", health, "ok\n")
		}

		metrics := httpGet(t, client, baseURL+"/metrics", http.StatusOK)
		for _, want := range []string{
			"template_exporter_build_info",
			`version="` + smokeVersion + `"`,
			`branch="` + smokeBranch + `"`,
			`revision="` + smokeRevision + `"`,
		} {
			if !strings.Contains(metrics, want) {
				t.Fatalf("GET /metrics body missing %q", want)
			}
		}
	})
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root from working directory")
		}
		dir = parent
	}
}

func buildBinary(t *testing.T, root string) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "prometheus-template-exporter")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	ldflags := strings.Join([]string{
		"-s",
		"-w",
		"-X github.com/prometheus/common/version.Version=" + smokeVersion,
		"-X github.com/prometheus/common/version.Branch=" + smokeBranch,
		"-X github.com/prometheus/common/version.Revision=" + smokeRevision,
		"-X github.com/prometheus/common/version.BuildUser=" + smokeBuildUser,
		"-X github.com/prometheus/common/version.BuildDate=" + smokeBuildDate,
	}, " ")

	cmd := exec.Command(goCommand(), "build", "-trimpath", "-ldflags", ldflags, "-o", binary, "./cmd")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}

func runBinary(t *testing.T, binary string, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", binary, strings.Join(args, " "), err, output)
	}
	return string(output)
}

func freeAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free address: %v", err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func waitForHealth(t *testing.T, client *http.Client, baseURL string, waitCh <-chan error, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err, ok := <-waitCh:
			if ok {
				t.Fatalf("server exited before health check passed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
			t.Fatalf("server exited before health check passed\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
		default:
		}

		resp, err := client.Get(baseURL + "/healthz")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for /healthz\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
}

func httpGet(t *testing.T, client *http.Client, url string, wantStatus int) string {
	t.Helper()

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read GET %s response body: %v", url, err)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("GET %s status = %d, want %d; body: %s", url, resp.StatusCode, wantStatus, body)
	}
	return string(body)
}

func goCommand() string {
	if goEnv := os.Getenv("GO"); goEnv != "" {
		return goEnv
	}
	return "go"
}
