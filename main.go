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

var version = "0.3.0-alpha.1"

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

	startHidden := false
	for _, arg := range os.Args[1:] {
		if arg == "--minimized" || arg == "-minimized" {
			startHidden = true
			break
		}
	}

	app := NewApp(token, logger)

	err := wails.Run(&options.App{
		Title:            "NETWATCH — Local Network Intelligence",
		Width:            1200,
		Height:           740,
		MinWidth:         1024,
		MinHeight:        640,
		StartHidden:      startHidden,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 12, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       func(ctx context.Context) { app.shutdown() },
		Bind:             []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "c18b76df-94bf-42f3-a261-netwatch-desktop",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				logger.Info("second instance launched, restoring window")
				app.RestoreWindow()
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			Theme:                windows.SystemDefault,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
