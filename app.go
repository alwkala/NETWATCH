//go:build windows

package main

import (
	"context"
	"fmt"
	"log/slog"

	"netwatch/internal/api"
	"netwatch/internal/appdata"
	"netwatch/internal/engine"
	"netwatch/internal/netenv"
	"netwatch/internal/store"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App owns the engine and the loopback API for the lifetime of the window.
type App struct {
	ctx     context.Context
	token   string
	log     *slog.Logger
	st      *store.Store
	eng     *engine.Engine
	baseURL string
	dataDir string
	cancel  context.CancelFunc
}

func NewApp(token string, log *slog.Logger) *App { return &App{token: token, log: log} }

// Connection is what the UI needs to reach the engine.
type Connection struct {
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
}

// GetConnection is called by the frontend (window.go.main.App.GetConnection).
func (a *App) GetConnection() Connection { return Connection{BaseURL: a.baseURL, Token: a.token} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	wruntime.WindowCenter(ctx)
	if err := a.start(ctx); err != nil {
		a.log.Error("engine failed to start", "err", err)
	}
}

// RestoreWindow brings the application window to foreground and restores if minimized.
func (a *App) RestoreWindow() {
	if a.ctx != nil {
		wruntime.WindowShow(a.ctx)
		wruntime.WindowUnminimise(a.ctx)
	}
}

func (a *App) start(parent context.Context) error {
	dir, err := appdata.Dir()
	if err != nil {
		return err
	}
	a.dataDir = dir
	st, err := store.Open(appdata.DBPath(dir))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	a.st = st
	ctx, cancel := context.WithCancel(parent)
	a.cancel = cancel

	eng := engine.New(engine.Options{Env: netenv.New(), Store: st, Logger: a.log})
	a.eng = eng
	eng.Start(ctx)

	srv := api.New(eng, api.Config{Token: a.token, Version: version, DBPath: appdata.DBPath(a.dataDir), Logger: a.log})
	srv.OpenDataFolder = func() error {
		return appdata.OpenFolder(a.dataDir)
	}
	srv.OpenURL = func(targetURL string) error {
		if a.ctx != nil {
			wruntime.BrowserOpenURL(a.ctx, targetURL)
			return nil
		}
		return appdata.OpenURL(targetURL)
	}
	ln, err := api.ListenLoopback(0)
	if err != nil {
		return err
	}
	a.baseURL = "http://" + ln.Addr().String()
	go func() {
		if err := srv.Serve(ctx, ln); err != nil {
			a.log.Error("api server", "err", err)
		}
	}()
	return nil
}

func (a *App) shutdown() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.eng != nil {
		a.eng.Stop()
	}
	if a.st != nil {
		a.st.Close()
	}
}
