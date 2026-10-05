//go:build windows

// NETWATCH desktop host: Wails (Go + WebView2) embeds the engine, serves the
// local API on 127.0.0.1 with a per-session token and shows the React UI.
package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"log"
	"log/slog"
	"os"

	"netwatch/internal/appdata"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "0.1.0-dev"

func main() {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(b[:])
	var (
		logger  *slog.Logger
		cleanup func()
	)
	if dir, err := appdata.Dir(); err == nil {
		logger, cleanup, _ = appdata.NewLogger(dir)
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	} else {
		defer cleanup()
	}
	app := NewApp(token, logger)

	err := wails.Run(&options.App{
		Title:       "NETWATCH",
		Width:       1440,
		Height:      900,
		MinWidth:    1180,
		MinHeight:   720,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		OnShutdown:  func(ctx context.Context) { app.shutdown() },
		Bind:        []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			Theme:                windows.SystemDefault,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
