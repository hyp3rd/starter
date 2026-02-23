// Package server provides the HTTP server implementation.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hyp3rd/ewrap"
)

const (
	// defaultPort is the default port if not set in env.
	defaultPort = "8000"
	// maxPort is the maximum valid port number.
	maxPort = 65535
	// probeTimeout is the timeout for health probes.
	probeTimeout = 5 * time.Second
	// DefaultVersion is the default app version if not set in env.
	DefaultVersion = "dev"
	// defaultSeparator is the default separator between host and port.
	defaultSeparator = ":"
)

// ErrProbeFailed indicates the health probe returned a non-200 status.
var ErrProbeFailed = errors.New("health probe failed with non-200 status")

// AddrFromEnv returns the listen address built from HOSTNAME and PORT envs.
func AddrFromEnv() string {
	host := os.Getenv("HOSTNAME")
	port := os.Getenv("PORT")

	if port == "" {
		port = defaultPort
	}

	if host == "" {
		return defaultSeparator + port
	}

	return host + defaultSeparator + port
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
	target, err := buildProbeTarget(ctx, listenAddr)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: probeTimeout,
	}

	return executeProbeRequest(ctx, target, &client, logger)
}

func buildProbeTarget(ctx context.Context, listenAddr string) (*url.URL, error) {
	host, port, err := splitListenAddress(listenAddr)
	if err != nil {
		return nil, err
	}

	err = validateProbeHost(ctx, host)
	if err != nil {
		return nil, err
	}

	normalizedPort, err := normalizePort(port)
	if err != nil {
		return nil, err
	}

	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", normalizedPort),
		Path:   "/health",
	}, nil
}

func splitListenAddress(listenAddr string) (host, port string, err error) {
	if strings.Contains(listenAddr, "://") {
		return "", "", ewrap.New("listen address must not include a URL scheme")
	}

	if after, ok := strings.CutPrefix(listenAddr, defaultSeparator); ok {
		return "", after, nil
	}

	host, port, err = net.SplitHostPort(listenAddr)
	if err != nil {
		return "", "", ewrap.Wrap(err, "invalid listen address")
	}

	return host, port, nil
}

func validateProbeHost(ctx context.Context, host string) error {
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" || strings.EqualFold(host, "localhost") {
		return nil
	}

	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return ewrap.Wrap(err, "failed to resolve health probe hostname")
	}

	if len(resolved) == 0 {
		return ewrap.Newf("no IP addresses found for hostname: %s", host)
	}

	for _, ipAddr := range resolved {
		if !isAllowedIP(ipAddr.IP) {
			return ewrap.Newf("health probe hostname resolves to disallowed address: %s", ipAddr.IP)
		}
	}

	return nil
}

func normalizePort(port string) (string, error) {
	if port == "" {
		return "", ewrap.New("listen port is empty")
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return "", ewrap.Wrap(err, "invalid listen port")
	}

	if parsedPort < 1 || parsedPort > maxPort {
		return "", ewrap.Newf("listen port out of range: %d", parsedPort)
	}

	return strconv.Itoa(parsedPort), nil
}

func executeProbeRequest(ctx context.Context, target *url.URL, client *http.Client, logger *slog.Logger) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return ewrap.Wrap(err, "failed to create health probe request")
	}
	// #nosec G704 -- target is canonicalized to loopback-only in buildProbeTarget
	resp, err := client.Do(req)
	if err != nil {
		return ewrap.Wrap(err, "health probe failed")
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

// isAllowedIP checks if an IP address is allowed for health probes.
// Only allows localhost/127.0.0.1 to prevent SSRF attacks.
func isAllowedIP(ip net.IP) bool {
	// Only allow loopback addresses (127.0.0.1 and ::1)
	return ip.IsLoopback()
}
