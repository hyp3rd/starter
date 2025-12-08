package server_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hyp3rd/starter/internal/server"
)

func provideLogger(t *testing.T) *slog.Logger {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return logger
}

func TestHealth(t *testing.T) {
	t.Setenv("APP_VERSION", "test-version")

	app := server.New()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error on body close: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAddrFromEnv(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("HOSTNAME", "0.0.0.0")

	addr := server.AddrFromEnv()
	if addr != "0.0.0.0:9000" {
		t.Fatalf("unexpected address: %s", addr)
	}

	t.Setenv("HOSTNAME", "")

	addr = server.AddrFromEnv()
	if addr != ":9000" {
		t.Fatalf("unexpected default host address: %s", addr)
	}
}

func TestVersionFromEnv(t *testing.T) {
	t.Setenv("APP_VERSION", "1.2.3")

	if got := server.VersionFromEnv(); got != "1.2.3" {
		t.Fatalf("unexpected version: %s", got)
	}

	err := os.Unsetenv("APP_VERSION")
	if err != nil {
		t.Fatal(err)
	}

	if got := server.VersionFromEnv(); got != server.DefaultVersion {
		t.Fatalf("expected default version, got: %s", got)
	}
}

func TestProbe(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	addr := strings.TrimPrefix(ts.URL, "http://")

	logger := provideLogger(t)

	err := server.Probe(context.TODO(), addr, logger)
	if err != nil {
		t.Fatalf("expected probe to succeed, got: %v", err)
	}
}

func TestProbeFails(t *testing.T) {
	t.Parallel()

	logger := provideLogger(t)

	err := server.Probe(context.TODO(), "127.0.0.1:0", logger)
	if err == nil {
		t.Fatal("expected probe to fail")
	}
}
