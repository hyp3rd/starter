// Package main implements the entry point for the application.
package main

import (
	"log/slog"
	"os"

	"github.com/hyp3rd/starter/internal/server"
)

func main() {
	app := server.New()
	addr := server.AddrFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("Starting server", slog.String("addr", addr))

	err := app.Listen(addr)
	if err != nil {
		logger.With("error", err).Error("failed to start server")

		return
	}
}
