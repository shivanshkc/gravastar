package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivanshkc/gravastar/internal/config"
	"github.com/shivanshkc/gravastar/internal/handlers"
	"github.com/shivanshkc/gravastar/internal/http"
	"github.com/shivanshkc/gravastar/internal/logger"
	"github.com/shivanshkc/gravastar/pkg/physics"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Allow the user to maintain multiple config files and switch between them conveniently.
	configPath := flag.String("config", "config/config.json", "config file path")
	flag.Parse()

	// Prime dependency.
	conf, err := config.Load(*configPath)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Set up the global logger. It can be used through slog.
	logger.Init(os.Stdout, conf.Logger.Level, conf.Logger.Pretty)

	// Initialize the HTTP server.
	engine := physics.NewGravityEngine(800, 600)
	go func() {
		slog.InfoContext(ctx, "starting gravity engine")
		engine.Run(ctx, conf.TargetFPS)
		slog.InfoContext(ctx, "gravity engine stopped")
	}()

	server := &http.Server{Handler: &handlers.Handler{Engine: engine}}

	// Start the http server. The server will shut down when the context expires.
	if err := server.Start(ctx, conf.HttpServer.Addr); err != nil {
		panic("error in server.Start call: " + err.Error())
	}
}
