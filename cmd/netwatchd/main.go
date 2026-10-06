// Command netwatchd runs the NETWATCH engine headless: the same engine and
// HTTP API the desktop app embeds, useful for development and as a future
// CLI/service. It prints one JSON line with the address and token on startup.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"netwatch/internal/api"
	"netwatch/internal/appdata"
	"netwatch/internal/engine"
	"netwatch/internal/netenv"
	"netwatch/internal/store"
)

var version = "0.2.0-alpha.1"

func main() {
	var (
		addrPort = flag.Int("port", 0, "TCP port on 127.0.0.1 (0 = random)")
		token    = flag.String("token", "", "bearer token (default: random per run)")
		dataDir  = flag.String("data", "", "data directory (default: per-user app data)")
		origins  = flag.String("origin", "", "extra allowed browser origins, comma separated (e.g. http://localhost:3000 for the Vite dev server)")
	)
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log, *addrPort, *token, *dataDir, *origins); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, port int, token, dataDir, origins string) error {
	if dataDir == "" {
		d, err := appdata.Dir()
		if err != nil {
			return err
		}
		dataDir = d
	}

	logger, cleanup, err := appdata.NewLogger(dataDir)
	if err == nil {
		defer cleanup()
		log = logger
	}

	if token == "" {
		var b [24]byte
		if _, err := rand.Read(b[:]); err != nil {
			return err
		}
		token = hex.EncodeToString(b[:])
	}
	var extra []string
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			n, err := api.ParseOrigin(o)
			if err != nil {
				return err
			}
			extra = append(extra, n)
		}
	}

	st, err := store.Open(appdata.DBPath(dataDir))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer st.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eng := engine.New(engine.Options{Env: netenv.New(), Store: st, Logger: log})
	eng.Start(ctx)

	srv := api.New(eng, api.Config{Token: token, AllowedOrigins: extra, Version: version, DBPath: appdata.DBPath(dataDir), Logger: log})
	srv.OpenDataFolder = func() error {
		return appdata.OpenFolder(dataDir)
	}
	ln, err := api.ListenLoopback(port)
	if err != nil {
		return err
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]string{
		"baseUrl": "http://" + ln.Addr().String(), "token": token, "data": dataDir,
	})
	log.Info("engine ready", "addr", ln.Addr().String(), "data", dataDir)
	return srv.Serve(ctx, ln)
}
