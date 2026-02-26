// Package server provides the HTTP server implementation.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	// defaultPort is the default port if not set in env.
	defaultPort = "8000"
	// probeTimeout is the timeout for health probes.
	probeTimeout = 5 * time.Second
	// DefaultVersion is the default app version if not set in env.
	DefaultVersion = "dev"
)

// ErrProbeFailed indicates the health probe returned a non-200 status.
var ErrProbeFailed = errors.New("health probe failed with non-200 status")

// ErrInvalidProbeTarget indicates the health probe target is not a local address.
var ErrInvalidProbeTarget = errors.New("health probe target must be local")

// AddrFromEnv returns the listen address built from HOSTNAME and PORT envs.
func AddrFromEnv() string {
	host := os.Getenv("HOSTNAME")
	port := os.Getenv("PORT")

	if port == "" {
		port = defaultPort
	}

	if host == "" {
		return ":" + port
	}

	return host + ":" + port
}

// VersionFromEnv returns the app version from APP_VERSION env or default.
func VersionFromEnv() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		return DefaultVersion
	}

	return version
}

// New constructs the Fiber application with default routes.
func New() *fiber.App {
	app := fiber.New()

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": VersionFromEnv(),
		})
	})

	return app
}

// Probe performs an HTTP GET against /health on the provided listen address.
func Probe(ctx context.Context, listenAddr string, logger *slog.Logger) error {
	target, err := probeTargetFromListenAddr(listenAddr)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: probeTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			// A health probe should not follow redirects to another location.
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("failed to create health probe request: %w", err)
	}

	// #nosec G704 -- target is validated as local-only and redirects are disabled
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health probe failed: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return ErrProbeFailed
	}

	return nil
}

func probeTargetFromListenAddr(listenAddr string) (string, error) {
	target := listenAddr
	if strings.HasPrefix(target, ":") {
		target = "127.0.0.1" + target
	}

	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return "", fmt.Errorf("failed to parse health probe target: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%w: unsupported scheme %q", ErrInvalidProbeTarget, parsed.Scheme)
	}

	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("%w: missing host", ErrInvalidProbeTarget)
	}

	normalizedHost, err := normalizeProbeHost(host)
	if err != nil {
		return "", err
	}

	port := parsed.Port()
	if port != "" {
		parsed.Host = net.JoinHostPort(normalizedHost, port)
	} else {
		parsed.Host = normalizedHost
	}

	parsed.Path = "/health"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func normalizeProbeHost(host string) (string, error) {
	if strings.EqualFold(host, "localhost") {
		return "localhost", nil
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("%w: %q", ErrInvalidProbeTarget, host)
	}

	if ip.IsUnspecified() {
		if ip.To4() != nil {
			return "127.0.0.1", nil
		}

		return "::1", nil
	}

	if ip.IsLoopback() {
		return ip.String(), nil
	}

	return "", fmt.Errorf("%w: %q", ErrInvalidProbeTarget, host)
}
