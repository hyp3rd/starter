// Package server provides the HTTP server implementation.
package server

import (
	"os"

	"github.com/gofiber/fiber/v3"
)

const (
	// defaultPort is the default port if not set in env.
	defaultPort = "8000"
	// DefaultVersion is the default app version if not set in env.
	DefaultVersion = "dev"
)

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
