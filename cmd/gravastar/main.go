package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"github.com/shivanshkc/gravastar/internal/config"
	"github.com/shivanshkc/gravastar/internal/handlers"
	httpx "github.com/shivanshkc/gravastar/internal/http"
	"github.com/shivanshkc/gravastar/internal/logger"
	"github.com/shivanshkc/gravastar/pkg/connor"
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
	engine := physics.NewGravityEngine(1000, 1000)
	// Engine cleanup.
	go physics.RightWallCollisionRemover(ctx, engine)

	go func() {
		slog.InfoContext(ctx, "starting gravity engine")
		engine.Run(ctx, conf.TargetFPS)
		slog.InfoContext(ctx, "gravity engine stopped")
	}()

	// Upgrader for websocket connections.
	upgrader := &websocket.Upgrader{
		HandshakeTimeout: time.Second * 10,
		CheckOrigin:      func(*http.Request) bool { return true },
	}

	// Websocket connection manager.
	manager := connor.NewManager(conf.WebsocketMaxConn, time.Second*30, slog.Default())

	// Update frontend config.
	if err := updateFrontendConfig(conf.HttpServer.StaticDir, conf.HttpServer.PublicAddr); err != nil {
		panic("failed to update frontend config: " + err.Error())
	}

	// HTTP + websocket server.
	server, err := httpx.NewServer(conf.HttpServer.Addr, conf.HttpServer.StaticDir,
		handlers.NewHandler(engine, upgrader, manager))
	if err != nil {
		panic("failed to create server: " + err.Error())
	}

	// Start the http server. The server will shut down when the context expires.
	if err := server.Start(ctx); err != nil {
		panic("error in server.Start call: " + err.Error())
	}
}

// updateFrontendConfig updates the config in the frontend code.
//
// It should be called before any files are served.
func updateFrontendConfig(staticDir, publicAddr string) error {
	const defaultConfig = "http://localhost:8080"
	filePath := path.Join(staticDir, "src/back.js")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read back.js: %w", err)
	}

	updatedContent := strings.ReplaceAll(string(content), defaultConfig, publicAddr)
	if err := os.WriteFile(filePath, []byte(updatedContent), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write back.js: %w", err)
	}

	return nil
}
