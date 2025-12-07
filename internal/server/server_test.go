package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hyp3rd/starter/internal/server"
)

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
