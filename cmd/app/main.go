// package main implements the entrypoint for the starter application.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/hyp3rd/starter/internal/server"
)

func main() {
	addrFlag := flag.String("addr", "", "listen address in host:port form (overrides HOSTNAME/PORT env)")
	healthcheck := flag.Bool("healthcheck", false, "run a local health probe against /health and exit")

	flag.Parse()

	addr := server.AddrFromEnv()
	if *addrFlag != "" {
		addr = *addrFlag
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *healthcheck {
		err := server.Probe(ctx, addr, logger)
		if err != nil {
			logger.With(
				slog.String("addr", addr),
				slog.Any("error", err),
			).Error("health probe failed")

			return
		}

		logger.With(slog.String("addr", addr)).Info("health probe succeeded")

		return
	}

	app := server.New()

	logger.Info("Starting server", slog.String("addr", addr))

	err := app.Listen(addr)
	if err != nil {
		logger.With("error", err).Error("failed to start server")
	}
}
